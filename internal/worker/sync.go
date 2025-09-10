package worker

import (
	"context"
	"encoding/json"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"
	"integration-education-db/internal/queue"
	"integration-education-db/internal/sync"

	"github.com/rs/zerolog"
)

type SyncWorker struct {
	cfg         *config.Config
	repo        db.Repository
	syncService *sync.Service
	consumer    *queue.Consumer
	workerPool  *WorkerPool
	log         zerolog.Logger
}

func NewSyncWorker(
	cfg *config.Config,
	repo db.Repository,
	redisClient *queue.RedisClient,
) *SyncWorker {
	return &SyncWorker{
		cfg:         cfg,
		repo:        repo,
		syncService: sync.NewService(cfg, repo),
		consumer:    queue.NewConsumer(redisClient, cfg),
		workerPool:  NewWorkerPool(cfg.Workers.Sync.Count),
		log:         logger.Get(),
	}
}

func (w *SyncWorker) Start(ctx context.Context) error {
	w.log.Info().Msg("Starting sync worker")

	// Start worker pool
	w.workerPool.Start(ctx)

	// Start consuming messages
	return w.consumer.ConsumeSyncQueue(ctx, w.handleMessage)
}

func (w *SyncWorker) Stop() {
	w.log.Info().Msg("Stopping sync worker")
	w.workerPool.Stop()
}

func (w *SyncWorker) handleMessage(ctx context.Context, data []byte) error {
	var job model.SyncJob
	if err := json.Unmarshal(data, &job); err != nil {
		w.log.Error().Err(err).Msg("Failed to unmarshal sync job")
		return err
	}

	w.log.Info().
		Int64("file_id", job.FileID).
		Str("class", job.Class).
		Str("subject", job.Subject).
		Msg("Processing sync job")

	// Submit job to worker pool
	w.workerPool.Submit(func(ctx context.Context) error {
		return w.syncService.ProcessSyncJob(ctx, job)
	})

	return nil
}
