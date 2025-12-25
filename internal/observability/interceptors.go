package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcLoggingUnaryInterceptor(otel.Tracer("grpc")),
		),
		grpc.ChainStreamInterceptor(
			grpcLoggingStreamInterceptor(otel.Tracer("grpc")),
		),
	}
}

type HTTPMiddleware func(http.Handler) http.Handler

func HTTPMiddlewareChain() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := ensureRequestID(requestIDFromHTTP(r))
			ctx := contextWithRequestID(r.Context(), requestID)
			spanCtx, span := otel.Tracer("http").Start(ctx, r.Method+" "+r.URL.Path)
			traceID := span.SpanContext().TraceID().String()
			r = r.WithContext(spanCtx)
			w.Header().Set("X-Request-Id", requestID)

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int64("http.status_code", int64(wrapped.status)),
				attribute.String("request.id", requestID),
			)
			span.End()

			logger := Logger()
			logEventForHTTP(logger, wrapped.status).
				Str("request_id", requestID).
				Str("trace_id", traceID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.status).
				Dur("duration", time.Since(start)).
				Msg("http request")
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

func grpcLoggingUnaryInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		ctx = propagateMetadata(ctx)
		requestID := ensureRequestID(requestIDFromMetadata(metadataFromContext(ctx)))
		ctx = contextWithRequestID(ctx, requestID)
		ctx, span := tracer.Start(ctx, info.FullMethod)
		_ = grpc.SetHeader(ctx, metadata.Pairs(requestIDHeader, requestID))
		span.SetAttributes(attribute.String("request.id", requestID))
		resp, err := handler(ctx, req)
		st := status.Convert(err)
		traceID := span.SpanContext().TraceID().String()
		logger := Logger()
		logEventForGRPC(logger, st.Code()).
			Str("request_id", requestID).
			Str("trace_id", traceID).
			Str("method", info.FullMethod).
			Str("code", st.Code().String()).
			Dur("duration", time.Since(start)).
			Msg("grpc unary")
		span.End()
		return resp, err
	}
}

func grpcLoggingStreamInterceptor(tracer trace.Tracer) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := propagateMetadata(stream.Context())
		requestID := ensureRequestID(requestIDFromMetadata(metadataFromContext(ctx)))
		ctx = contextWithRequestID(ctx, requestID)
		ctx, span := tracer.Start(ctx, info.FullMethod)
		_ = stream.SetHeader(metadata.Pairs(requestIDHeader, requestID))
		span.SetAttributes(attribute.String("request.id", requestID))
		wrapped := &wrappedServerStream{ServerStream: stream, ctx: ctx}
		err := handler(srv, wrapped)
		st := status.Convert(err)
		traceID := span.SpanContext().TraceID().String()
		logger := Logger()
		logEventForGRPC(logger, st.Code()).
			Str("request_id", requestID).
			Str("trace_id", traceID).
			Str("method", info.FullMethod).
			Str("code", st.Code().String()).
			Dur("duration", time.Since(start)).
			Msg("grpc stream")
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

func metadataFromContext(ctx context.Context) metadata.MD {
	md, _ := metadata.FromIncomingContext(ctx)
	return md
}

func logEventForHTTP(logger zerolog.Logger, status int) *zerolog.Event {
	switch {
	case status >= http.StatusInternalServerError:
		return logger.Error()
	case status >= http.StatusBadRequest:
		return logger.Warn()
	default:
		return logger.Info()
	}
}

func logEventForGRPC(logger zerolog.Logger, code codes.Code) *zerolog.Event {
	switch code {
	case codes.OK:
		return logger.Info()
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition, codes.Unauthenticated, codes.PermissionDenied:
		return logger.Warn()
	default:
		return logger.Error()
	}
}
