package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Dheeraj2209/Last_mile_go/internal/storage"
)

type Config struct {
	ServiceName    string
	GRPCListenAddr string
	GRPCEndpoint   string
	HTTPAddr       string
	OTelEndpoint   string
	OTelInsecure   bool
	LogLevel       string

	Mongo storage.MongoConfig
	Redis storage.RedisConfig
}

func Load(serviceName string) Config {
	return Config{
		ServiceName:    serviceName,
		GRPCListenAddr: getEnv("GRPC_LISTEN_ADDR", ":9090"),
		GRPCEndpoint:   getEnv("GRPC_ENDPOINT", "localhost:9090"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		OTelEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTelInsecure:   getEnvBool("OTEL_EXPORTER_OTLP_INSECURE", true),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		Mongo: storage.MongoConfig{
			URI:     os.Getenv("MONGO_URI"),
			Timeout: getEnvDuration("MONGO_TIMEOUT", 10*time.Second),
		},
		Redis: storage.RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB", 0),
			Timeout:  getEnvDuration("REDIS_TIMEOUT", 5*time.Second),
		},
	}
}

func Validate(cfg Config) error {
	var errs []error
	if cfg.ServiceName == "" {
		errs = append(errs, errors.New("SERVICE_NAME is required"))
	}
	if cfg.GRPCListenAddr == "" {
		errs = append(errs, errors.New("GRPC_LISTEN_ADDR is required"))
	}
	if cfg.GRPCEndpoint == "" {
		errs = append(errs, errors.New("GRPC_ENDPOINT is required"))
	}
	if cfg.HTTPAddr == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}
	if cfg.Mongo.URI == "" {
		// optional for now; no validation
	}
	if cfg.Redis.Addr == "" {
		// optional for now; no validation
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func FormatConfig(cfg Config) string {
	return fmt.Sprintf("grpc_listen=%s grpc_endpoint=%s http_addr=%s otel_endpoint=%s otel_insecure=%t log_level=%s mongo_uri_set=%t redis_addr_set=%t",
		cfg.GRPCListenAddr,
		cfg.GRPCEndpoint,
		cfg.HTTPAddr,
		cfg.OTelEndpoint,
		cfg.OTelInsecure,
		cfg.LogLevel,
		cfg.Mongo.URI != "",
		cfg.Redis.Addr != "",
	)
}
