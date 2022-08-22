package log

import (
	"context"
)

type contextKey string

const loggerKey = contextKey("logger")

// FromCtx returns Logger instance from the provided context. If no such logger exists,
// no-op logger is provided.
func FromCtx(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l
	}
	return NewNoopLogger()
}

// WithCtx creates a new context with the provided Logger attached.
func WithCtx(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
