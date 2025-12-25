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
	"github.com/Dheeraj2209/Last_mile_go/services/matching"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load("matching")

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownOTel, err := observability.Setup(ctx, cfg.ServiceName, cfg.OTelEndpoint, cfg.OTelInsecure)
	if err != nil {
		log.Fatalf("matching service failed to init telemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			log.Printf("matching telemetry shutdown error: %v", err)
		}
	}()

	err = server.Run(ctx, cfg.GRPCListenAddr, cfg.GRPCEndpoint, cfg.HTTPAddr,
		func(grpcServer *grpc.Server) {
			lastmilev1.RegisterMatchingServiceServer(grpcServer, matching.NewServer())
		},
		lastmilev1.RegisterMatchingServiceHandlerFromEndpoint,
	)
	if err != nil {
		log.Fatalf("matching service stopped: %v", err)
	}
}
