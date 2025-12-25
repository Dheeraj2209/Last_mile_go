package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Dheeraj2209/Last_mile_go/internal/observability"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCRegistrar func(*grpc.Server)

type GatewayRegistrar func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

func Run(ctx context.Context, grpcListenAddr, grpcEndpoint, httpAddr string, registerGRPC GRPCRegistrar, registerGateway GatewayRegistrar) error {
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(observability.GRPCServerOptions()...)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)
	healthgrpc.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	registerGRPC(grpcServer)

	errCh := make(chan error, 2)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	gatewayMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	registerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := registerGateway(registerCtx, gatewayMux, grpcEndpoint, dialOpts); err != nil {
		grpcServer.Stop()
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", observability.HTTPMiddlewareChain()(gatewayMux))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
	return nil
}
