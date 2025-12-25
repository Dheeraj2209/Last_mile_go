package storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Timeout  time.Duration
}

func NewRedisClient(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, errors.New("redis addr is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
