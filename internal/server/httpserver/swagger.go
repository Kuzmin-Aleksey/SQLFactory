package httpserver

import (
	_ "embed"
	"net/http"
	"strconv"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//go:embed openapi.yaml
var openapiSpec []byte

func (s *Server) serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")

	// Some proxies/healthchecks use HEAD requests.
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.Itoa(len(openapiSpec)))
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiSpec)
}

func (s *Server) swaggerUIHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	)
}
