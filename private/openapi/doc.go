package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ScalarHandler returns an http.Handler that serves a Swagger UI page using CDN-hosted Scalar UI.
func ScalarHandler(docs []*openapi3.T, specPathPrefix string, serverURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine base URL dynamically if serverURL not set
		currentServerURL := serverURL
		if currentServerURL == "" {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			currentServerURL = fmt.Sprintf("%s://%s", scheme, r.Host)
		}

		// If requesting raw spec json
		if r.URL.Path == specPathPrefix+"spec.json" || r.URL.Path == specPathPrefix+"spec.json/" {
			w.Header().Set("Content-Type", "application/json")
			if len(docs) == 0 {
				_ = json.NewEncoder(w).Encode(map[string]any{"openapi": "3.0.0", "info": map[string]any{"title": "Empty Spec", "version": "1.0"}})
				return
			}

			var targetDoc *openapi3.T
			if len(docs) == 1 {
				// Copy doc to avoid mutating original
				docCopy := *docs[0]
				targetDoc = &docCopy
			} else {
				// Merge multiple docs
				components := openapi3.NewComponents()
				targetDoc = &openapi3.T{
					OpenAPI: docs[0].OpenAPI,
					Info: &openapi3.Info{
						Title:       "FauxRPC Merged OpenAPI Specification",
						Version:     "1.0.0",
						Description: "Combined documentation of loaded OpenAPI schemas",
					},
					Paths:      openapi3.NewPaths(),
					Components: &components,
				}
				for _, doc := range docs {
					if doc == nil {
						continue
					}
					if doc.Paths != nil {
						for path, item := range doc.Paths.Map() {
							targetDoc.Paths.Set(path, item)
						}
					}
					mergeComponents(targetDoc.Components, doc.Components)
				}
			}

			// Add the local FauxRPC server as the primary server while preserving
			// the schema's base path (for example, /api/v3).
			localServerURL := currentServerURL
			var serverVariables map[string]*openapi3.ServerVariable
			if len(docs) == 1 && len(docs[0].Servers) > 0 {
				sourceServer := docs[0].Servers[0]
				localServerURL = withServerBasePath(currentServerURL, sourceServer.URL)
				serverVariables = sourceServer.Variables
			}
			serverObj := &openapi3.Server{
				URL:         localServerURL,
				Description: "FauxRPC Server",
				Variables:   serverVariables,
			}
			targetDoc.Servers = append(openapi3.Servers{serverObj}, targetDoc.Servers...)

			_ = json.NewEncoder(w).Encode(targetDoc)
			return
		}

		specURL := specPathPrefix + "spec.json"

		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <title>FauxRPC OpenAPI Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body {
        margin: 0;
      }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>
`, specURL)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
}

func mergeComponents(target, source *openapi3.Components) {
	if target == nil || source == nil {
		return
	}
	mergeComponentMap(&target.Extensions, source.Extensions)
	mergeComponentMap(&target.Schemas, source.Schemas)
	mergeComponentMap(&target.Parameters, source.Parameters)
	mergeComponentMap(&target.Headers, source.Headers)
	mergeComponentMap(&target.RequestBodies, source.RequestBodies)
	mergeComponentMap(&target.Responses, source.Responses)
	mergeComponentMap(&target.SecuritySchemes, source.SecuritySchemes)
	mergeComponentMap(&target.Examples, source.Examples)
	mergeComponentMap(&target.Links, source.Links)
	mergeComponentMap(&target.Callbacks, source.Callbacks)
}

func mergeComponentMap[M ~map[string]V, V any](target *M, source M) {
	if len(source) == 0 {
		return
	}
	if *target == nil {
		*target = make(M, len(source))
	}
	for name, value := range source {
		(*target)[name] = value
	}
}

func withServerBasePath(serverURL, schemaServerURL string) string {
	parsed, err := url.Parse(schemaServerURL)
	if err != nil || parsed.Path == "" || parsed.Path == "/" {
		return strings.TrimRight(serverURL, "/")
	}
	return strings.TrimRight(serverURL, "/") + "/" + strings.TrimLeft(parsed.Path, "/")
}
