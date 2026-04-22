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

	rtr.HandleFunc("/api/template", s.MwAuth(s.saveTemplate)).Methods(http.MethodPost)
	rtr.HandleFunc("/api/template", s.MwAuth(s.updateTemplate)).Methods(http.MethodPut)
	rtr.HandleFunc("/api/template", s.MwAuth(s.deleteTemplate)).Methods(http.MethodDelete)
	rtr.HandleFunc("/api/templates", s.MwAuth(s.getDBTemplates)).Methods(http.MethodGet)

	rtr.HandleFunc("/api/history", s.MwAuth(s.getDBHistory)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/history/item", s.MwAuth(s.getHistoryItem)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/history", s.MwAuth(s.deleteHistoryItem)).Methods(http.MethodDelete)

	rtr.HandleFunc("/api/dict", s.MwAuth(s.getDBDict)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.addDictItem)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.updateDictItem)).Methods(http.MethodPut)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.deleteDictItem)).Methods(http.MethodDelete)

	rtr.HandleFunc("/api/executor/prompt", s.MwAuth(s.executeUserPrompt))
	rtr.HandleFunc("/api/executor/template", s.MwAuth(s.executeTemplate))
	rtr.HandleFunc("/api/executor/history", s.MwAuth(s.executeHistory))
}
