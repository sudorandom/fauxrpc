package openapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsOpenAPISpec(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. OpenAPI YAML file
	openapiYaml := filepath.Join(tmpDir, "api.yaml")
	err := os.WriteFile(openapiYaml, []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
`), 0644)
	assert.NoError(t, err)

	assert.True(t, IsOpenAPISpec(openapiYaml))

	// 2. Swagger JSON file
	swaggerJson := filepath.Join(tmpDir, "swagger.json")
	err = os.WriteFile(swaggerJson, []byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Legacy API",
    "version": "1.0.0"
  }
}`), 0644)
	assert.NoError(t, err)

	assert.True(t, IsOpenAPISpec(swaggerJson))

	// 3. Protobuf / Non-OpenAPI file
	nonOpenapi := filepath.Join(tmpDir, "service.json")
	err = os.WriteFile(nonOpenapi, []byte(`{
  "name": "something_else"
}`), 0644)
	assert.NoError(t, err)

	assert.False(t, IsOpenAPISpec(nonOpenapi))
}

func TestIsOpenAPISpecClosesNonOKResponseBody(t *testing.T) {
	body := &trackingBody{Reader: strings.NewReader("not found")}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       body,
			Header:     make(http.Header),
		}, nil
	})}

	assert.False(t, isOpenAPISpec("https://example.invalid/openapi.yaml", client))
	assert.True(t, body.closed)
}

func TestIsOpenAPISpecURLTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	started := time.Now()
	assert.False(t, isOpenAPISpec(server.URL, &http.Client{Timeout: 20 * time.Millisecond}))
	assert.Less(t, time.Since(started), time.Second)
}

type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
