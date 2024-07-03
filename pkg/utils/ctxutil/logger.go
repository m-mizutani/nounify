package ctxutil

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/nounify/pkg/utils/logging"
)

type ctxLoggerKey struct{}

func Logger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger)
	if !ok {
		return logging.Default()
	}
	return logger
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}
