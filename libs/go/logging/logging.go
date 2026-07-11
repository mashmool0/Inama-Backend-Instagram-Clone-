// Package logging gives every service the same structured JSON logger.
// One format everywhere → logs are greppable and ready for a log aggregator.
package logging

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is unexported so only this package can set the trace id.
type contextKey struct{}

var traceIDKey = contextKey{}

// New returns a JSON logger tagged with the service name. LOG_LEVEL controls
// verbosity (debug|info|warn|error), defaulting to info.
//
//	log := logging.New("user")
//	log.Info("server started", "port", 8080)
func New(service string) *slog.Logger {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With("service", service)
}

// WithTraceID stashes a trace id on the context so downstream logging can pick
// it up. Wired to OpenTelemetry later — the field is reserved now on purpose.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// FromContext returns a logger that automatically includes the context's trace
// id (if present), so every log line in a request is correlated.
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if tid, ok := ctx.Value(traceIDKey).(string); ok && tid != "" {
		return base.With("trace_id", tid)
	}
	return base
}
