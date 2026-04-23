package httpserver

import (
	_ "embed"
	"io"
	"net/http"
	"strconv"
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

func (s *Server) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodHead {
		return
	}

	_, _ = io.WriteString(w, swaggerUIPage)
}

const swaggerUIPage = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SQLFactory API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" crossorigin="anonymous">
  <style>body { margin: 0; } #swagger-ui { max-width: 100%; }</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin="anonymous"></script>
<script>
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: window.location.origin + '/openapi.yaml',
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis],
    layout: 'BaseLayout'
  });
};
</script>
</body>
</html>
`
