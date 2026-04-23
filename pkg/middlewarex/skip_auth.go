package middlewarex

import (
	"SQLFactory/pkg/contextx"
	"net/http"
)

func NewScipAuthMw(userId contextx.UserId) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = contextx.WithUserId(ctx, userId)
			ctx = contextx.WithSkipAuth(ctx, true)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
