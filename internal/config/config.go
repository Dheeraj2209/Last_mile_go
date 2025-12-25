package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServiceName    string
	GRPCListenAddr string
	GRPCEndpoint   string
	HTTPAddr       string
	OTelEndpoint   string
	OTelInsecure   bool
}

func Load(serviceName string) Config {
	return Config{
		ServiceName:    serviceName,
		GRPCListenAddr: getEnv("GRPC_LISTEN_ADDR", ":9090"),
		GRPCEndpoint:   getEnv("GRPC_ENDPOINT", "localhost:9090"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		OTelEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTelInsecure:   getEnvBool("OTEL_EXPORTER_OTLP_INSECURE", true),
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

func FormatConfig(cfg Config) string {
	return fmt.Sprintf("grpc_listen=%s grpc_endpoint=%s http_addr=%s otel_endpoint=%s otel_insecure=%t",
		cfg.GRPCListenAddr,
		cfg.GRPCEndpoint,
		cfg.HTTPAddr,
		cfg.OTelEndpoint,
		cfg.OTelInsecure,
	)
}
