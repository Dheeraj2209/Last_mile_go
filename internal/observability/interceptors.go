package observability

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Logger interface {
	Printf(format string, v ...any)
}

func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcLoggingUnaryInterceptor(otel.Tracer("grpc"), defaultLogger{}),
		),
		grpc.ChainStreamInterceptor(
			grpcLoggingStreamInterceptor(otel.Tracer("grpc"), defaultLogger{}),
		),
	}
}

type HTTPMiddleware func(http.Handler) http.Handler

func HTTPMiddlewareChain() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			spanCtx, span := otel.Tracer("http").Start(r.Context(), r.Method+" "+r.URL.Path)
			r = r.WithContext(spanCtx)

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int64("http.status_code", int64(wrapped.status)),
			)
			span.End()

			defaultLogger{}.Printf("http request method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, wrapped.status, time.Since(start))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type defaultLogger struct{}

func (defaultLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func grpcLoggingUnaryInterceptor(tracer trace.Tracer, logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		ctx = propagateMetadata(ctx)
		ctx, span := tracer.Start(ctx, info.FullMethod)
		resp, err := handler(ctx, req)
		st := status.Convert(err)
		logger.Printf("grpc unary method=%s code=%s duration=%s", info.FullMethod, st.Code().String(), time.Since(start))
		span.End()
		return resp, err
	}
}

func grpcLoggingStreamInterceptor(tracer trace.Tracer, logger Logger) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := propagateMetadata(stream.Context())
		ctx, span := tracer.Start(ctx, info.FullMethod)
		wrapped := &wrappedServerStream{ServerStream: stream, ctx: ctx}
		err := handler(srv, wrapped)
		st := status.Convert(err)
		logger.Printf("grpc stream method=%s code=%s duration=%s", info.FullMethod, st.Code().String(), time.Since(start))
		span.End()
		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func propagateMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	return metadata.NewIncomingContext(ctx, md)
}
