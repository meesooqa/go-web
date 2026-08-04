package lgr

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// NewContext returns a context with an attached logger. It is useful for
// passing a request-scoped logger (for example, with a request_id) through
// HTTP middleware or a gRPC interceptor.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext retrieves the logger from the context. If no logger was
// attached via NewContext, it returns slog.Default(),
// the calling code does not need to check for nil.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
