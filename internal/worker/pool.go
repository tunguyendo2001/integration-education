package worker

import (
	"context"
	"sync"

	"integration-education-db/internal/logger"

	"github.com/rs/zerolog"
)

type WorkerPool struct {
	workerCount int
	jobChan     chan func(context.Context) error
	wg          sync.WaitGroup
	log         zerolog.Logger
}

func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobChan:     make(chan func(context.Context) error, workerCount*2),
		log:         logger.Get(),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	wp.log.Info().Int("worker_count", wp.workerCount).Msg("Starting worker pool")

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) Stop() {
	wp.log.Info().Msg("Stopping worker pool")
	close(wp.jobChan)
	wp.wg.Wait()
	wp.log.Info().Msg("Worker pool stopped")
}

func (wp *WorkerPool) Submit(job func(context.Context) error) {
	select {
	case wp.jobChan <- job:
	default:
		wp.log.Warn().Msg("Worker pool job queue full, job dropped")
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	log := wp.log.With().Int("worker_id", id).Logger()
	log.Debug().Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Worker stopping due to context cancellation")
			return
		case job, ok := <-wp.jobChan:
			if !ok {
				log.Debug().Msg("Worker stopping due to closed job channel")
				return
			}

			if err := job(ctx); err != nil {
				log.Error().Err(err).Msg("Job execution failed")
			}
		}
	}
}
