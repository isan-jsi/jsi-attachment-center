package api

import (
	"net/http"

	topapi "github.com/jsi/ibs-doc-engine/api"
)

// ServeOpenAPISpec returns a handler that serves the raw OpenAPI YAML spec.
func ServeOpenAPISpec() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(topapi.OpenAPISpec)
	}
}

// SwaggerUIHTML returns a minimal HTML page that loads Swagger UI from CDN.
func SwaggerUIHTML(specURL string) http.HandlerFunc {
	page := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>IBS Doc Engine API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "` + specURL + `",
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	}
}
