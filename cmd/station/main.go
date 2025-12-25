package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/Dheeraj2209/Last_mile_go/internal/config"
	"github.com/Dheeraj2209/Last_mile_go/internal/observability"
	"github.com/Dheeraj2209/Last_mile_go/internal/server"
	"github.com/Dheeraj2209/Last_mile_go/services/station"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load("station")

	var grpcListenAddr string
	var grpcEndpoint string
	var httpAddr string
	var otelEndpoint string
	var otelInsecure bool

	flag.StringVar(&grpcListenAddr, "grpc-listen", cfg.GRPCListenAddr, "gRPC listen address")
	flag.StringVar(&grpcEndpoint, "grpc-endpoint", cfg.GRPCEndpoint, "gRPC endpoint for gateway dialing")
	flag.StringVar(&httpAddr, "http-addr", cfg.HTTPAddr, "HTTP listen address")
	flag.StringVar(&otelEndpoint, "otel-endpoint", cfg.OTelEndpoint, "OTel OTLP gRPC endpoint (host:port)")
	flag.BoolVar(&otelInsecure, "otel-insecure", cfg.OTelInsecure, "Disable TLS for OTLP exporter")
	flag.Parse()

	cfg.GRPCListenAddr = grpcListenAddr
	cfg.GRPCEndpoint = grpcEndpoint
	cfg.HTTPAddr = httpAddr
	cfg.OTelEndpoint = otelEndpoint
	cfg.OTelInsecure = otelInsecure

	if err := config.Validate(cfg); err != nil {
		log.Fatalf("station invalid configuration: %v", err)
	}
	log.Printf("station config: %s", config.FormatConfig(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownOTel, err := observability.Setup(ctx, cfg.ServiceName, cfg.OTelEndpoint, cfg.OTelInsecure)
	if err != nil {
		log.Fatalf("station service failed to init telemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			log.Printf("station telemetry shutdown error: %v", err)
		}
	}()

	err = server.Run(ctx, cfg.GRPCListenAddr, cfg.GRPCEndpoint, cfg.HTTPAddr,
		func(grpcServer *grpc.Server) {
			lastmilev1.RegisterStationServiceServer(grpcServer, station.NewServer())
		},
		lastmilev1.RegisterStationServiceHandlerFromEndpoint,
	)
	if err != nil {
		log.Fatalf("station service stopped: %v", err)
	}
}
