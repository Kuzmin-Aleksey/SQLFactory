package app

import (
	"SQLFactory/pkg/contextx"
	"net/http"
)

type serverDisableAuth struct {
	restServerInterface
}

func newServerDisableAuth(server restServerInterface) *serverDisableAuth {
	return &serverDisableAuth{
		restServerInterface: server,
	}
}

func (s *serverDisableAuth) MwAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = contextx.WithUserId(ctx, contextx.UserId(debugUserId))
		r = r.WithContext(ctx)
		next(w, r)
	}
}
