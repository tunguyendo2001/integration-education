package queue

import (
	"context"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/logger"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

type Consumer struct {
	client *redis.Client
	cfg    *config.Config
	log    zerolog.Logger
}

type MessageHandler func(ctx context.Context, data []byte) error

func NewConsumer(redisClient *RedisClient, cfg *config.Config) *Consumer {
	return &Consumer{
		client: redisClient.Client(),
		cfg:    cfg,
		log:    logger.Get(),
	}
}

func (c *Consumer) ConsumeIngestionQueue(ctx context.Context, handler MessageHandler) error {
	return c.consume(ctx, c.cfg.Redis.IngestionQueue, handler)
}

func (c *Consumer) ConsumeSyncQueue(ctx context.Context, handler MessageHandler) error {
	return c.consume(ctx, c.cfg.Redis.SyncQueue, handler)
}

func (c *Consumer) consume(ctx context.Context, queueName string, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result, err := c.client.BRPop(ctx, 5*time.Second, queueName).Result()
			if err != nil {
				if err == redis.Nil {
					continue // Timeout, continue polling
				}
				c.log.Error().Err(err).Str("queue", queueName).Msg("Failed to consume message")
				continue
			}

			if len(result) < 2 {
				continue
			}

			message := result[1]
			if err := handler(ctx, []byte(message)); err != nil {
				c.log.Error().Err(err).Str("queue", queueName).Msg("Failed to process message")
				// Move to DLQ
				dlqName := queueName + c.cfg.Redis.DLQSuffix
				if dlqErr := c.client.LPush(ctx, dlqName, message).Err(); dlqErr != nil {
					c.log.Error().Err(dlqErr).Str("dlq", dlqName).Msg("Failed to move message to DLQ")
				}
			}
		}
	}
}
