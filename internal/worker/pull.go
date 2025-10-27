package worker

import (
	"context"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/pull"

	"github.com/rs/zerolog"
)

type PullWorker struct {
	cfg         *config.Config
	repo        db.Repository
	pullService *pull.Service
	timer       *time.Timer
	log         zerolog.Logger
}

func NewPullWorker(
	cfg *config.Config,
	repo db.Repository,
) *PullWorker {
	return &PullWorker{
		cfg:         cfg,
		repo:        repo,
		pullService: pull.NewService(cfg, repo),
		log:         logger.Get(),
	}
}

func (w *PullWorker) Start(ctx context.Context) error {
	w.log.Info().Msg("Starting pull worker")

	// Calculate next run time (end of day)
	nextRun := w.getNextRunTime()
	w.log.Info().Time("next_run", nextRun).Msg("Scheduled next pull")

	// Pull immediately on start (optional - you can comment this out)
	if w.cfg.Workers.Pull.RunOnStart {
		w.log.Info().Msg("Running initial pull on startup")
		if err := w.pullAll(ctx); err != nil {
			w.log.Error().Err(err).Msg("Initial pull failed")
		}
	}

	// Setup ticker for daily pulls at end of day
	duration := time.Until(nextRun)
	w.timer = time.NewTimer(duration)

	for {
		select {
		case <-ctx.Done():
			w.log.Info().Msg("Pull worker context cancelled")
			return ctx.Err()
		case <-w.timer.C:
			w.log.Info().Msg("Starting scheduled pull")
			if err := w.pullAll(ctx); err != nil {
				w.log.Error().Err(err).Msg("Scheduled pull failed")
			}

			// Schedule next run (end of next day)
			nextRun = w.getNextRunTime()
			w.log.Info().Time("next_run", nextRun).Msg("Scheduled next pull")
			w.timer.Reset(time.Until(nextRun))
		}
	}
}

func (w *PullWorker) Stop() {
	w.log.Info().Msg("Stopping pull worker")
	if w.timer != nil {
		w.timer.Stop()
	}
}

// getNextRunTime returns the time for end of current day
func (w *PullWorker) getNextRunTime() time.Time {
	now := time.Now()

	// Get end of current day (23:59:59)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// If we're past end of day, move to next day
	if now.After(endOfDay) {
		endOfDay = endOfDay.Add(24 * time.Hour)
	}

	return endOfDay
}

func (w *PullWorker) pullAll(ctx context.Context) error {
	startTime := time.Now()
	w.log.Info().Msg("Starting data pull from server")

	var teacherCount, studentCount, classCount int
	var hasErrors bool

	// Pull teachers
	count, err := w.pullService.PullTeachers(ctx)
	teacherCount = count
	if err != nil {
		w.log.Error().Err(err).Msg("Failed to pull teachers")
		hasErrors = true
	} else {
		w.log.Info().Int("count", teacherCount).Msg("Teachers pulled successfully")
	}

	// Pull students
	count, err = w.pullService.PullStudents(ctx)
	studentCount = count
	if err != nil {
		w.log.Error().Err(err).Msg("Failed to pull students")
		hasErrors = true
	} else {
		w.log.Info().Int("count", studentCount).Msg("Students pulled successfully")
	}

	// Pull classes
	count, err = w.pullService.PullClasses(ctx)
	classCount = count
	if err != nil {
		w.log.Error().Err(err).Msg("Failed to pull classes")
		hasErrors = true
	} else {
		w.log.Info().Int("count", classCount).Msg("Classes pulled successfully")
	}

	duration := time.Since(startTime)
	w.log.Info().
		Dur("duration", duration).
		Int("teachers", teacherCount).
		Int("students", studentCount).
		Int("classes", classCount).
		Bool("has_errors", hasErrors).
		Msg("Data pull completed")

	if hasErrors {
		return context.DeadlineExceeded // Return error but don't crash
	}

	return nil
}
