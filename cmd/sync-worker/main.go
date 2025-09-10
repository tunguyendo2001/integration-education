package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/queue"
	"integration-education-db/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log := logger.Get()

	log.Info().Str("version", cfg.App.Version).Msg("Starting sync worker")

	// Initialize database
	database, err := db.NewConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// Initialize repository
	repo := db.NewRepository(database)

	// Initialize Redis client
	redisClient, err := queue.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Create sync worker
	syncWorker := worker.NewSyncWorker(cfg, repo, redisClient)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	go func() {
		if err := syncWorker.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Sync worker failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down sync worker...")

	// Cancel context to stop worker
	cancel()
	syncWorker.Stop()

	log.Info().Msg("Sync worker exited")
}
