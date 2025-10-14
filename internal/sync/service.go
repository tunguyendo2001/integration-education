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
	cfg               *config.Config
	repo              db.Repository
	client            *Client
	permissionChecker *PermissionChecker
	log               zerolog.Logger
}

func NewService(cfg *config.Config, repo db.Repository) *Service {
	return &Service{
		cfg:               cfg,
		repo:              repo,
		client:            NewClient(cfg),
		permissionChecker: NewPermissionChecker(cfg),
		log:               logger.Get(),
	}
}

// CheckSyncPermission checks if sync is allowed via API endpoint
func (s *Service) CheckSyncPermission(ctx context.Context, className string, semester string, year int) (bool, error) {
	s.log.Debug().
		Str("class", className).
		Str("semester", semester).
		Int("year", year).
		Msg("Checking sync permission via API")

	isAllowed, resp, err := s.permissionChecker.CheckPermission(ctx, className, semester, year)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to check permission")
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	if !isAllowed {
		schedule := resp.GetActiveSchedule()
		if schedule != nil {
			if schedule.Locked {
				s.log.Warn().
					Str("class", className).
					Str("semester", semester).
					Int("year", year).
					Msg("Schedule is locked")
				return false, nil
			}

			if !schedule.Active {
				s.log.Warn().
					Str("class", className).
					Str("semester", semester).
					Int("year", year).
					Msg("Schedule is not active")
				return false, nil
			}

			// Check if current time is within schedule window
			now := time.Now()
			if now.Before(schedule.StartDateTime) || now.After(schedule.EndDateTime) {
				s.log.Warn().
					Str("class", className).
					Time("now", now).
					Time("start", schedule.StartDateTime).
					Time("end", schedule.EndDateTime).
					Msg("Current time is outside schedule window")
				return false, nil
			}
		} else {
			s.log.Warn().
				Str("class", className).
				Str("semester", semester).
				Int("year", year).
				Msg("No active schedule found")
			return false, nil
		}
	}

	s.log.Info().
		Str("class", className).
		Str("semester", semester).
		Int("year", year).
		Msg("Sync permission granted")

	return true, nil
}

func (s *Service) ProcessSyncJob(ctx context.Context, job model.SyncJob) error {
	log := s.log.With().
		Int64("file_id", job.FileID).
		Str("class", job.Class).
		Str("subject", job.Subject).
		Logger()

	log.Info().Msg("Processing sync job")

	// Extract year and semester from job (you might need to get this from file or grades)
	// For now, assuming you have a way to determine these
	// You might need to add Year and Semester fields to SyncJob
	year := job.Year
	semester := job.Semester

	// Check sync permission via API
	allowed, err := s.CheckSyncPermission(ctx, job.Class, semester, year)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check sync permission")
		return err
	}

	if !allowed {
		log.Warn().Msg("Sync not allowed for this class/semester/year")
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
