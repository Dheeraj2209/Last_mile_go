package config

import (
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
	cfg := Config{
		ServiceName:    serviceName,
		GRPCListenAddr: getEnv("GRPC_LISTEN_ADDR", ":9090"),
		GRPCEndpoint:   getEnv("GRPC_ENDPOINT", "localhost:9090"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		OTelEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTelInsecure:   getEnvBool("OTEL_EXPORTER_OTLP_INSECURE", true),
	}
	return cfg
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
