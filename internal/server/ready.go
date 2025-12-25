package server

import (
	"context"
	"fmt"
	"log"

	"github.com/Dheeraj2209/Last_mile_go/internal/config"
	"github.com/Dheeraj2209/Last_mile_go/internal/storage"
)

type ReadyCheck func(context.Context) error

type ReadyResources struct {
	Checks  []ReadyCheck
	Closers []func(context.Context) error
}

func ReadyChecksFromConfig(ctx context.Context, cfg config.Config, logf func(string, ...any)) (ReadyResources, error) {
	res := ReadyResources{}
	logger := logf
	if logger == nil {
		logger = log.Printf
	}

	if cfg.Mongo.URI != "" {
		client, err := storage.NewMongoClient(ctx, cfg.Mongo)
		if err != nil {
			return res, fmt.Errorf("mongo readiness init: %w", err)
		}
		res.Checks = append(res.Checks, func(ctx context.Context) error {
			return client.Ping(ctx, nil)
		})
		res.Closers = append(res.Closers, client.Disconnect)
		logger("readiness: mongo enabled")
	}

	if cfg.Redis.Addr != "" {
		client, err := storage.NewRedisClient(ctx, cfg.Redis)
		if err != nil {
			return res, fmt.Errorf("redis readiness init: %w", err)
		}
		res.Checks = append(res.Checks, func(ctx context.Context) error {
			return client.Ping(ctx).Err()
		})
		res.Closers = append(res.Closers, func(context.Context) error {
			return client.Close()
		})
		logger("readiness: redis enabled")
	}

	return res, nil
}
