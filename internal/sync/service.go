package sync

import (
	"context"
	"fmt"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"
	"integration-education-db/pkg/errors"

	"github.com/rs/zerolog"
)

type Service struct {
	cfg    *config.Config
	repo   db.Repository
	client *Client
	log    zerolog.Logger
}

func NewService(cfg *config.Config, repo db.Repository) *Service {
	return &Service{
		cfg:    cfg,
		repo:   repo,
		client: NewClient(cfg),
		log:    logger.Get(),
	}
}

func (s *Service) IsWithinSyncWindow() (bool, error) {
	now := time.Now()
	location, err := time.LoadLocation(s.cfg.SyncWindow.Timezone)
	if err != nil {
		return false, fmt.Errorf("invalid timezone: %w", err)
	}

	nowInTZ := now.In(location)

	startTime, err := time.Parse("15:04", s.cfg.SyncWindow.StartTime)
	if err != nil {
		return false, fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("15:04", s.cfg.SyncWindow.EndTime)
	if err != nil {
		return false, fmt.Errorf("invalid end time format: %w", err)
	}

	currentTime := nowInTZ.Format("15:04")
	startTimeStr := startTime.Format("15:04")
	endTimeStr := endTime.Format("15:04")

	return currentTime >= startTimeStr && currentTime <= endTimeStr, nil
}

func (s *Service) ProcessSyncJob(ctx context.Context, job model.SyncJob) error {
	log := s.log.With().
		Int64("file_id", job.FileID).
		Str("class", job.Class).
		Str("subject", job.Subject).
		Logger()

	log.Info().Msg("Processing sync job")

	// Check sync window
	withinWindow, err := s.IsWithinSyncWindow()
	if err != nil {
		log.Error().Err(err).Msg("Failed to check sync window")
		return err
	}

	if !withinWindow {
		log.Warn().Msg("Sync attempted outside allowed window")
		return errors.ErrSyncWindowClosed
	}

	// Get ready grades in batches
	batchSize := s.cfg.ExternalAPI.EducationDept.BatchSize
	offset := 0
	totalProcessed := 0

	for {
		grades, err := s.repo.GetReadyGrades(ctx, job.FileID, job.Class, job.Subject, batchSize)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get ready grades")
			return err
		}

		if len(grades) == 0 {
			break // No more grades to process
		}

		log.Debug().Int("batch_size", len(grades)).Int("offset", offset).Msg("Processing batch")

		if err := s.processBatch(ctx, grades); err != nil {
			log.Error().Err(err).Msg("Failed to process batch")
			return err
		}

		totalProcessed += len(grades)
		offset += batchSize

		// If we got less than batch size, we're done
		if len(grades) < batchSize {
			break
		}
	}

	log.Info().Int("total_processed", totalProcessed).Msg("Sync job completed")
	return nil
}

func (s *Service) processBatch(ctx context.Context, grades []model.GradeStaging) error {
	if len(grades) == 0 {
		return nil
	}

	// Collect IDs for batch updates
	gradeIDs := make([]int64, len(grades))
	for i, grade := range grades {
		gradeIDs[i] = grade.ID
	}

	// Send to external API with retries
	var lastErr error
	for attempt := 0; attempt < s.cfg.ExternalAPI.EducationDept.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.cfg.ExternalAPI.EducationDept.RetryDelay * time.Duration(attempt+1)):
				// Exponential backoff
			}
		}

		resp, err := s.client.SendGradeBatch(ctx, grades)
		if err != nil {
			lastErr = err
			s.log.Warn().Err(err).Int("attempt", attempt+1).Msg("Batch send failed, retrying")
			continue
		}

		// Process response
		if resp.Success {
			// All grades succeeded
			return s.repo.UpdateGradesStatus(ctx, gradeIDs, model.GradeStatusSynced, nil)
		} else {
			// Some grades failed - handle partial success
			return s.handlePartialSuccess(ctx, grades, resp)
		}
	}

	// All retries exhausted
	errorMsg := fmt.Sprintf("Max retries exhausted: %v", lastErr)
	return s.repo.UpdateGradesStatus(ctx, gradeIDs, model.GradeStatusFailed, &errorMsg)
}

func (s *Service) handlePartialSuccess(ctx context.Context, grades []model.GradeStaging, resp *model.BatchResponse) error {
	// Create a map of failed student IDs
	failedStudents := make(map[string]bool)
	for _, studentID := range resp.Failed {
		failedStudents[studentID] = true
	}

	var successIDs, failedIDs []int64
	for _, grade := range grades {
		if failedStudents[grade.StudentID] {
			failedIDs = append(failedIDs, grade.ID)
		} else {
			successIDs = append(successIDs, grade.ID)
		}
	}

	// Update successful records
	if len(successIDs) > 0 {
		if err := s.repo.UpdateGradesStatus(ctx, successIDs, model.GradeStatusSynced, nil); err != nil {
			return err
		}
	}

	// Update failed records
	if len(failedIDs) > 0 {
		errorMsg := resp.Message
		if errorMsg == "" {
			errorMsg = "External API reported failure"
		}
		if err := s.repo.UpdateGradesStatus(ctx, failedIDs, model.GradeStatusFailed, &errorMsg); err != nil {
			return err
		}
	}

	return nil
}
