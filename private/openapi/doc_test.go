package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarHandlerPreservesServerBasePath(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Petstore",
			Version: "1.0.0",
		},
		Servers: openapi3.Servers{
			{URL: "/api/v3"},
		},
		Paths: openapi3.NewPaths(),
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/fauxrpc/openapi-docs/spec.json", nil)
	ScalarHandler(
		[]*openapi3.T{doc},
		"/fauxrpc/openapi-docs/",
		"http://127.0.0.1:6660",
	).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var servedDoc openapi3.T
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&servedDoc))
	require.NotEmpty(t, servedDoc.Servers)
	assert.Equal(t, "http://127.0.0.1:6660/api/v3", servedDoc.Servers[0].URL)
	assert.Equal(t, "/api/v3", doc.Servers[0].URL, "source document must not be mutated")
}

func TestWithServerBasePathFromAbsoluteURL(t *testing.T) {
	assert.Equal(
		t,
		"http://localhost:6660/api/v3",
		withServerBasePath("http://localhost:6660/", "https://petstore.example/api/v3"),
	)
}

func TestScalarHandlerUsesValidDoctype(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/fauxrpc/openapi-docs/", nil)
	ScalarHandler(nil, "/fauxrpc/openapi-docs/", "").ServeHTTP(recorder, request)

	assert.NotContains(t, recorder.Body.String(), "<!utf-8>")
	assert.Contains(t, recorder.Body.String(), "<!DOCTYPE html>")
}

func TestScalarHandlerMergesComponents(t *testing.T) {
	loader := openapi3.NewLoader()
	users, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info: {title: Users, version: 1.0.0}
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: {$ref: '#/components/schemas/User'}
components:
  schemas:
    User:
      type: object
      properties:
        name: {type: string}
`))
	require.NoError(t, err)
	orders, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info: {title: Orders, version: 1.0.0}
paths:
  /orders:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: {$ref: '#/components/schemas/Order'}
components:
  schemas:
    Order:
      type: object
      properties:
        id: {type: integer}
`))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/fauxrpc/openapi-docs/spec.json", nil)
	ScalarHandler(
		[]*openapi3.T{users, orders},
		"/fauxrpc/openapi-docs/",
		"http://127.0.0.1:6660",
	).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var servedDoc openapi3.T
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&servedDoc))
	require.NotNil(t, servedDoc.Components)
	assert.Contains(t, servedDoc.Components.Schemas, "User")
	assert.Contains(t, servedDoc.Components.Schemas, "Order")
	assert.NotNil(t, servedDoc.Paths.Value("/users"))
	assert.NotNil(t, servedDoc.Paths.Value("/orders"))
}
