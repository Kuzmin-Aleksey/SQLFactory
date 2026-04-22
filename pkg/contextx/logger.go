package contextx

import (
	"context"
	"log/slog"
	"os"
)

type contextKeyLogger struct{}

var DefaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger{}, logger)
}

func GetLoggerOrDefault(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextKeyLogger{}).(*slog.Logger)
	if !ok || logger == nil {
		return DefaultLogger
	}
	return logger
}
