package httpserver

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) RegisterRoutes(rtr *mux.Router) {
	rtr.HandleFunc("/api/auth/register", s.Register).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/login", s.Login).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/refresh", s.Refresh).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/check-token", s.MwAuth(func(http.ResponseWriter, *http.Request) {}))
}
