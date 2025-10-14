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

	log.Info().Str("version", cfg.App.Version).Msg("Starting pull worker")

	// Initialize database
	database, err := db.NewConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// Initialize repository
	repo := db.NewRepository(database)

	// Create pull worker
	pullWorker := worker.NewPullWorker(cfg, repo)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	go func() {
		if err := pullWorker.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Pull worker failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down pull worker...")

	// Cancel context to stop worker
	cancel()
	pullWorker.Stop()

	log.Info().Msg("Pull worker exited")
}
