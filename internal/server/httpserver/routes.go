package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes(rtr *mux.Router) {
	// API docs
	rtr.HandleFunc("/openapi.yaml", s.serveOpenAPI).Methods(http.MethodGet, http.MethodHead)
	rtr.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	}).Methods(http.MethodGet, http.MethodHead)
	rtr.PathPrefix("/swagger/").Handler(s.swaggerUIHandler())
	rtr.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusFound)
	}).Methods(http.MethodGet, http.MethodHead)

	rtr.HandleFunc("/api/auth/register", s.Register).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/login", s.Login).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/refresh", s.Refresh).Methods(http.MethodPost)
	rtr.HandleFunc("/api/auth/check-token", s.MwAuth(func(http.ResponseWriter, *http.Request) {})).Methods(http.MethodGet)

	rtr.HandleFunc("/api/template", s.MwAuth(s.getTemplate)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/template", s.MwAuth(s.saveTemplate)).Methods(http.MethodPost)
	rtr.HandleFunc("/api/template", s.MwAuth(s.updateTemplate)).Methods(http.MethodPut)
	rtr.HandleFunc("/api/template", s.MwAuth(s.deleteTemplate)).Methods(http.MethodDelete)
	rtr.HandleFunc("/api/templates", s.MwAuth(s.getDBTemplates)).Methods(http.MethodGet)

	rtr.HandleFunc("/api/history", s.MwAuth(s.getDBHistory)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/history/item", s.MwAuth(s.getHistoryItem)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/history/items", s.MwAuth(s.getItemsByFirstId)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/history", s.MwAuth(s.deleteHistoryItem)).Methods(http.MethodDelete)

	rtr.HandleFunc("/api/dict", s.MwAuth(s.getDBDict)).Methods(http.MethodGet)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.addDictItem)).Methods(http.MethodPost)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.updateDictItem)).Methods(http.MethodPut)
	rtr.HandleFunc("/api/dict/item", s.MwAuth(s.deleteDictItem)).Methods(http.MethodDelete)

	rtr.HandleFunc("/api/executor/test-connect", s.MwAuth(s.testConnect))
	rtr.HandleFunc("/api/executor/prompt", s.MwAuth(s.executeUserPrompt))
	rtr.HandleFunc("/api/executor/template", s.MwAuth(s.executeTemplate))
	rtr.HandleFunc("/api/executor/history", s.MwAuth(s.executeHistory))
}
