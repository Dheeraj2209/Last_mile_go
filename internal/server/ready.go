package server

import (
	"context"
	"fmt"
	"log"

	"github.com/Dheeraj2209/Last_mile_go/internal/config"
	"github.com/Dheeraj2209/Last_mile_go/internal/storage"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReadyCheck func(context.Context) error

type ReadyResources struct {
	Checks  []ReadyCheck
	Closers []func(context.Context) error
}

func ReadyChecksFromConfig(ctx context.Context, cfg config.Config, logf func(string, ...any)) (ReadyResources, error) {
	var mongoClient *mongo.Client
	var redisClient *redis.Client

	if cfg.Mongo.URI != "" {
		client, err := storage.NewMongoClient(ctx, cfg.Mongo)
		if err != nil {
			return ReadyResources{}, fmt.Errorf("mongo readiness init: %w", err)
		}
		mongoClient = client
	}

	if cfg.Redis.Addr != "" {
		client, err := storage.NewRedisClient(ctx, cfg.Redis)
		if err != nil {
			return ReadyResources{}, fmt.Errorf("redis readiness init: %w", err)
		}
		redisClient = client
	}

	return ReadyChecksFromClients(mongoClient, redisClient, logf), nil
}

func ReadyChecksFromClients(mongoClient *mongo.Client, redisClient *redis.Client, logf func(string, ...any)) ReadyResources {
	res := ReadyResources{}
	logger := logf
	if logger == nil {
		logger = log.Printf
	}
	if mongoClient != nil {
		res.Checks = append(res.Checks, func(ctx context.Context) error {
			return mongoClient.Ping(ctx, nil)
		})
		res.Closers = append(res.Closers, mongoClient.Disconnect)
		logger("readiness: mongo enabled")
	}
	if redisClient != nil {
		res.Checks = append(res.Checks, func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		})
		res.Closers = append(res.Closers, func(context.Context) error {
			return redisClient.Close()
		})
		logger("readiness: redis enabled")
	}
	return res
}
