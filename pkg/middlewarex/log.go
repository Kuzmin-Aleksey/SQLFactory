package middlewarex

import (
	"SQLFactory/internal/util"
	"SQLFactory/pkg/contextx"
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type LogOptions struct {
	MaxContentLen   int // < 0 log full content; == 0 not log
	SensitiveFields []string
	LoggingContent  []string
}

func NewLogRequest(opts *LogOptions) func(http.Handler) http.Handler {
	if opts == nil {
		opts = &LogOptions{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				logRequest(r, opts)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewLogResponse(opts *LogOptions) func(http.Handler) http.Handler {
	if opts == nil {
		opts = &LogOptions{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customWriter := &LoggingWriter{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
				MaxContentLen:  opts.MaxContentLen,
			}

			start := time.Now()
			next.ServeHTTP(customWriter, r)
			end := time.Now()
			if strings.HasPrefix(r.URL.Path, "/api/") {
				logResponse(r.Context(), customWriter, end.Sub(start), opts)
			}
		})
	}
}

type LoggingWriter struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
	Content       []byte
	MaxContentLen int
}

func (r *LoggingWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *LoggingWriter) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)

	if r.ContentLength+n < r.MaxContentLen || r.MaxContentLen < 0 {
		r.Content = append(r.Content, p...)
	} else if r.ContentLength < r.MaxContentLen {
		r.Content = append(r.Content, p[:r.MaxContentLen-r.ContentLength]...)
	}

	r.ContentLength += n

	return n, err
}

func logRequest(r *http.Request, opts *LogOptions) {
	ctx := r.Context()
	l := contextx.GetLoggerOrDefault(ctx)

	contentType := r.Header.Get("Content-Type")

	attrs := []slog.Attr{
		slog.String("protocol", r.Proto),
		slog.String("method", r.Method),
		slog.String("url", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("content_length", r.ContentLength),
		slog.String("content_type", contentType),
	}

	if r.Form == nil {
		r.ParseForm()
	}

	if len(r.Form) > 0 {
		attrs = append(attrs, slog.Group("values", getSafeSlogValues(r.Form, opts.SensitiveFields)...))
	}

	if slices.Contains(opts.LoggingContent, contentType) && r.ContentLength != 0 {
		if opts.MaxContentLen < 0 {
			opts.MaxContentLen = int(r.ContentLength)
		}

		body := make([]byte, min(r.ContentLength, int64(opts.MaxContentLen)))
		if _, err := r.Body.Read(body); err != nil && !errors.Is(err, io.EOF) {
			l.LogAttrs(ctx, slog.LevelError, "read request body error", slog.String("err", err.Error()))
		}

		if contentType == "application/json" {
			unmarshalledBody := jsonUnmarshal(body)
			hideSensitiveFields(unmarshalledBody, opts.SensitiveFields)

			attrs = append(attrs, slog.Any("content", unmarshalledBody))
		} else {
			attrs = append(attrs, slog.String("content", string(body)))
		}

		r.Body = util.NewMultiReadCloser(io.NopCloser(bytes.NewReader(body)), r.Body)
	}

	l.LogAttrs(ctx, slog.LevelInfo, "request received", attrs...)
}

func logResponse(ctx context.Context, r *LoggingWriter, handleDuration time.Duration, opts *LogOptions) {
	l := contextx.GetLoggerOrDefault(ctx)
	contentType := r.Header().Get("Content-Type")

	attrs := []slog.Attr{
		slog.String("status", strconv.Itoa(r.StatusCode)),
		slog.Duration("handle_duration", handleDuration),
		slog.String("content_type", contentType),
		slog.Int("content_len", r.ContentLength),
	}

	if slices.Contains(opts.LoggingContent, contentType) && len(r.Content) > 0 {
		if contentType == "application/json" {
			unmarshalledBody := jsonUnmarshal(r.Content)
			hideSensitiveFields(unmarshalledBody, opts.SensitiveFields)

			attrs = append(attrs, slog.Any("content", unmarshalledBody))
		} else {
			attrs = append(attrs, slog.String("content", string(r.Content)))
		}
	}

	l.LogAttrs(ctx, slog.LevelInfo, "response sent", attrs...)
}

func hideSensitiveFields(v any, fields []string) {
	switch t := v.(type) {
	case map[string]any:
		for k := range t {
			if slices.Contains(fields, strings.ToLower(k)) {
				t[k] = "hidden"
			} else {
				hideSensitiveFields(t[k], fields)
			}
		}
	case []any:
		for i := range t {
			hideSensitiveFields(t[i], fields)
		}
	}
}

func getSafeSlogValues(v url.Values, safeFields []string) []any {
	if v == nil {
		return nil
	}

	attrs := make([]any, 0, len(v))

	for k := range v {
		var val string
		if slices.Contains(safeFields, strings.ToLower(k)) {
			val = "hidden"
		} else {
			val = v.Get(k)
		}

		attrs = append(attrs, slog.String(k, val))
	}

	return attrs
}
