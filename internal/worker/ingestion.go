package worker

import (
	"context"
	"encoding/json"
	"io"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/excel"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"
	"integration-education-db/internal/queue"
	"integration-education-db/internal/storage"

	"github.com/rs/zerolog"
)

type IngestionWorker struct {
	cfg        *config.Config
	repo       db.Repository
	storage    storage.Storage
	parser     excel.ParsingStrategy
	consumer   *queue.Consumer
	workerPool *WorkerPool
	log        zerolog.Logger
}

func NewIngestionWorker(
	cfg *config.Config,
	repo db.Repository,
	storage storage.Storage,
	redisClient *queue.RedisClient,
) *IngestionWorker {
	return &IngestionWorker{
		cfg:        cfg,
		repo:       repo,
		storage:    storage,
		parser:     excel.NewExcelStrategy(),
		consumer:   queue.NewConsumer(redisClient, cfg),
		workerPool: NewWorkerPool(cfg.Workers.Ingestion.Count),
		log:        logger.Get(),
	}
}

func (w *IngestionWorker) Start(ctx context.Context) error {
	w.log.Info().Msg("Starting ingestion worker")

	// Start worker pool
	w.workerPool.Start(ctx)

	// Start consuming messages
	return w.consumer.ConsumeIngestionQueue(ctx, w.handleMessage)
}

func (w *IngestionWorker) Stop() {
	w.log.Info().Msg("Stopping ingestion worker")
	w.workerPool.Stop()
}

func (w *IngestionWorker) handleMessage(ctx context.Context, data []byte) error {
	var job model.IngestionJob
	if err := json.Unmarshal(data, &job); err != nil {
		w.log.Error().Err(err).Msg("Failed to unmarshal ingestion job")
		return err
	}

	w.log.Info().Int64("file_id", job.FileID).Str("s3_path", job.S3Path).Msg("Processing ingestion job")

	// Submit job to worker pool
	w.workerPool.Submit(func(ctx context.Context) error {
		return w.processFile(ctx, job)
	})

	return nil
}

func (w *IngestionWorker) processFile(ctx context.Context, job model.IngestionJob) error {
	log := w.log.With().Int64("file_id", job.FileID).Logger()

	// Download file from S3
	log.Debug().Msg("Downloading file from S3")
	reader, err := w.storage.Download(ctx, job.S3Path)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download file")
		errorMsg := err.Error()
		w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedFail, &errorMsg)
		return err
	}
	defer reader.Close()

	// Read file data
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read file data")
		errorMsg := err.Error()
		w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedFail, &errorMsg)
		return err
	}

	// Parse Excel file
	log.Debug().Msg("Parsing Excel file")
	grades, err := w.parser.Parse(ctx, data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse Excel file")
		errorMsg := err.Error()
		w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedFail, &errorMsg)
		return err
	}

	// Validate parsed data
	log.Debug().Int("grade_count", len(grades)).Msg("Validating parsed grades")
	if err := w.parser.Validate(ctx, grades); err != nil {
		log.Error().Err(err).Msg("Grade validation failed")
		errorMsg := err.Error()
		w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedFail, &errorMsg)
		return err
	}

	// Insert into staging table
	log.Debug().Msg("Inserting grades into staging table")
	if err := w.repo.InsertGrades(ctx, job.FileID, grades); err != nil {
		log.Error().Err(err).Msg("Failed to insert grades")
		errorMsg := err.Error()
		w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedFail, &errorMsg)
		return err
	}

	// Update file status to success
	if err := w.repo.UpdateFileStatus(ctx, job.FileID, model.FileStatusParsedOK, nil); err != nil {
		log.Error().Err(err).Msg("Failed to update file status")
		return err
	}

	log.Info().Int("grade_count", len(grades)).Msg("File processed successfully")
	return nil
}
