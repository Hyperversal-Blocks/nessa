package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(ctx context.Context, config *redis.Options) (*RedisClient, error) {
	client := redis.NewClient(config)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to bootstrap service: %w", err)
	}

	return &RedisClient{
		Client: client,
	}, nil
}
