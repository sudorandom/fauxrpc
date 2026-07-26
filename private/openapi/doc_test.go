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
