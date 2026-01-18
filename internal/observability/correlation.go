package observability

import (
	"context"

	"github.com/rs/zerolog"
)

func AttachTraceContext(ctx context.Context, logger *zerolog.Logger) *zerolog.Logger {
	return EnrichLogger(ctx, logger)
}

func CreateChildLogger(ctx context.Context, component string) *zerolog.Logger {
	base := LoggerFromContext(ctx)
	child := base.With().Str("component", component).Logger()
	return EnrichLogger(ctx, &child)
}
