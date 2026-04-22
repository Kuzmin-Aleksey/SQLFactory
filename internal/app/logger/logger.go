package logger

import (
	"SQLFactory/internal/config"
	"SQLFactory/pkg/contextx"
	"context"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	LogFormatJson = "json"
	LogFormatText = "text"
)

type logHandler struct {
	slog.Handler
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	traceId := contextx.GetTraceId(ctx)

	if traceId != "" {
		r.AddAttrs(slog.String("trace_id", string(traceId)))
	}

	return h.Handler.Handle(ctx, r)
}

func GetLogger(cfg *config.LogConfig) (*slog.Logger, error) {
	out := io.Writer(os.Stdout)

	if cfg.Path != "" {
		rt, err := rotatelogs.New(filepath.Join(cfg.Path, "%Y-%m-%d.log"),
			rotatelogs.WithRotationTime(time.Hour*24),
			rotatelogs.WithMaxAge(time.Hour*24*15),
		)
		if err != nil {
			log.Fatal(err)
		}
		out = io.MultiWriter(rt, os.Stdout)
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case LogFormatText:
		handler = slog.NewTextHandler(out, opts)
	case LogFormatJson:
		handler = slog.NewJSONHandler(out, opts)
	default:
		log.Println("unknown logging format ", cfg.Format)
		handler = slog.NewJSONHandler(out, nil)
	}

	l := slog.New(&logHandler{
		Handler: handler,
	})

	return l, nil

}
