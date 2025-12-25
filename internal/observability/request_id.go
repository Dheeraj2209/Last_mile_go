package observability

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const requestIDHeader = "x-request-id"

type requestIDKey struct{}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(requestIDKey{}).(string); ok {
		return value
	}
	return ""
}

func contextWithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

func ensureRequestID(id string) string {
	if strings.TrimSpace(id) != "" {
		return strings.TrimSpace(id)
	}
	return uuid.NewString()
}

func requestIDFromHTTP(r *http.Request) string {
	if r == nil {
		return ""
	}
	for _, key := range []string{
		"X-Request-Id",
		"X-Request-ID",
		"X-Correlation-Id",
		"X-Correlation-ID",
	} {
		if value := strings.TrimSpace(r.Header.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func requestIDFromMetadata(md metadata.MD) string {
	if md == nil {
		return ""
	}
	for _, key := range []string{requestIDHeader, "x-correlation-id"} {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		if value := strings.TrimSpace(values[0]); value != "" {
			return value
		}
	}
	return ""
}
