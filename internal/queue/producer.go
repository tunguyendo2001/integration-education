package queue

import (
	"context"
	"encoding/json"

	"integration-education-db/internal/config"
	"integration-education-db/internal/model"

	"github.com/go-redis/redis/v8"
)

type Producer struct {
	client *redis.Client
	cfg    *config.Config
}

func NewProducer(redisClient *RedisClient, cfg *config.Config) *Producer {
	return &Producer{
		client: redisClient.Client(),
		cfg:    cfg,
	}
}

func (p *Producer) EnqueueIngestionJob(ctx context.Context, job model.IngestionJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.client.LPush(ctx, p.cfg.Redis.IngestionQueue, data).Err()
}

func (p *Producer) EnqueueSyncJob(ctx context.Context, job model.SyncJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.client.LPush(ctx, p.cfg.Redis.SyncQueue, data).Err()
}
