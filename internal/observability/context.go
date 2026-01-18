package observability

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	loggerKey contextKey = "agk.logger"
	runIDKey  contextKey = "agk.run_id"
)

func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	if l, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok && l != nil {
		return l
	}

	nop := zerolog.Nop()
	return &nop
}

func WithRunID(ctx context.Context, runID string) context.Context {
	if runID == "" {
		return ctx
	}
	return context.WithValue(ctx, runIDKey, runID)
}

func RunIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(runIDKey).(string); ok {
		return id
	}
	return ""
}

func EnrichLogger(ctx context.Context, base *zerolog.Logger) *zerolog.Logger {
	if base == nil {
		nop := zerolog.Nop()
		base = &nop
	}

	sc := trace.SpanContextFromContext(ctx)
	l := base.With()

	if sc.IsValid() {
		l = l.
			Str("trace_id", sc.TraceID().String()).
			Str("span_id", sc.SpanID().String())
	}

	if runID := RunIDFromContext(ctx); runID != "" {
		l = l.Str("run_id", runID)
	}

	enriched := l.Logger()
	return &enriched
}
