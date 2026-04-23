package middlewarex

import (
	"SQLFactory/pkg/contextx"
	"github.com/google/uuid"
	"net/http"
)

const headerTraceId = "X-Trace-Id"

func AddTraceId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get(headerTraceId)
		if traceId == "" {
			traceId = uuid.NewString()
			w.Header().Set(headerTraceId, traceId)
		}

		ctx := contextx.WithTraceId(r.Context(), contextx.TraceId(traceId))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
