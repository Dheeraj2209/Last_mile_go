package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/Dheeraj2209/Last_mile_go/internal/config"
	"github.com/Dheeraj2209/Last_mile_go/internal/observability"
	"github.com/Dheeraj2209/Last_mile_go/internal/server"
	"github.com/Dheeraj2209/Last_mile_go/services/driver"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load("driver")

	var grpcListenAddr string
	var grpcEndpoint string
	var httpAddr string
	var otelEndpoint string
	var otelInsecure bool
	var logLevel string

	flag.StringVar(&grpcListenAddr, "grpc-listen", cfg.GRPCListenAddr, "gRPC listen address")
	flag.StringVar(&grpcEndpoint, "grpc-endpoint", cfg.GRPCEndpoint, "gRPC endpoint for gateway dialing")
	flag.StringVar(&httpAddr, "http-addr", cfg.HTTPAddr, "HTTP listen address")
	flag.StringVar(&otelEndpoint, "otel-endpoint", cfg.OTelEndpoint, "OTel OTLP gRPC endpoint (host:port)")
	flag.BoolVar(&otelInsecure, "otel-insecure", cfg.OTelInsecure, "Disable TLS for OTLP exporter")
	flag.StringVar(&logLevel, "log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")
	flag.Parse()

	cfg.GRPCListenAddr = grpcListenAddr
	cfg.GRPCEndpoint = grpcEndpoint
	cfg.HTTPAddr = httpAddr
	cfg.OTelEndpoint = otelEndpoint
	cfg.OTelInsecure = otelInsecure
	cfg.LogLevel = logLevel

	logger := observability.ConfigureLogger(cfg.ServiceName, cfg.LogLevel)
	if err := config.Validate(cfg); err != nil {
		logger.Fatal().Err(err).Msg("invalid configuration")
	}
	logger.Info().Str("config", config.FormatConfig(cfg)).Msg("service config")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownOTel, err := observability.Setup(ctx, cfg.ServiceName, cfg.OTelEndpoint, cfg.OTelInsecure)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init telemetry")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("telemetry shutdown error")
		}
	}()

	ready, err := server.ReadyChecksFromConfig(ctx, cfg, observability.Logf())
	if err != nil {
		logger.Fatal().Err(err).Msg("readiness init failed")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, closer := range ready.Closers {
			if closer == nil {
				continue
			}
			if err := closer(shutdownCtx); err != nil {
				logger.Error().Err(err).Msg("readiness close error")
			}
		}
	}()

	err = server.Run(ctx, cfg.GRPCListenAddr, cfg.GRPCEndpoint, cfg.HTTPAddr,
		func(grpcServer *grpc.Server) {
			lastmilev1.RegisterDriverServiceServer(grpcServer, driver.NewServer())
		},
		lastmilev1.RegisterDriverServiceHandlerFromEndpoint,
		ready.Checks...,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("service stopped")
	}
}
