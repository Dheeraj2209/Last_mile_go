package storage

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI     string
	Timeout time.Duration
}

func NewMongoClient(ctx context.Context, cfg MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, errors.New("mongo uri is required")
	}
	clientOpts := options.Client().ApplyURI(cfg.URI)
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
