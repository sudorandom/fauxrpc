package engine

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/openapi"
)

func TestEngineDispatcher(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users/{id}:
    get:
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: role
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: OK
          headers:
            X-Rate-Limit:
              schema:
                type: integer
                example: 250
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
`))
	assert.NoError(t, err)

	router, err := openapi.NewRouter(doc)
	assert.NoError(t, err)

	registry := newMockStubRegistry()
	registry.AddStub(StubRule{
		Name: "Explicit Rest Stub",
		Target: TargetRef{
			OperationID: "getUserById",
		},
		Match: StubMatch{
			PathParams: map[string]string{
				"id": "usr_rest_123",
			},
		},
		Response: StubResponse{
			Status: 200,
			Body: map[string]any{
				"id":   "usr_rest_123",
				"name": "REST Admin User",
			},
		},
	})

	dispatcher := NewDispatcher(registry, router, 5, false, false)

	// 1. Test stub match
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/v1/users/usr_rest_123", nil)
	handled := dispatcher.ServeHTTP(rec1, req1)
	assert.True(t, handled)
	assert.Equal(t, 200, rec1.Code)
	assert.Contains(t, rec1.Body.String(), "REST Admin User")

	// 2. Test auto-generator fallback
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/v1/users/usr_other", nil)
	handled2 := dispatcher.ServeHTTP(rec2, req2)
	assert.True(t, handled2)
	assert.Equal(t, 200, rec2.Code)
	assert.Equal(t, "250", rec2.Header().Get("X-Rate-Limit"))
	assert.Contains(t, rec2.Body.String(), "id")
}

func TestEngineDispatcherOnlyStubs(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users/{id}:
    get:
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
`))
	require.NoError(t, err)
	router, err := openapi.NewRouter(doc)
	require.NoError(t, err)

	registry := newMockStubRegistry()
	registry.AddStub(StubRule{
		Target: TargetRef{OperationID: "getUserById"},
		Match:  StubMatch{PathParams: map[string]string{"id": "stubbed"}},
		Response: StubResponse{
			Status: http.StatusOK,
			Body:   map[string]any{"id": "stubbed"},
		},
	})
	dispatcher := NewDispatcher(registry, router, 5, false, true)

	stubbed := httptest.NewRecorder()
	assert.True(t, dispatcher.ServeHTTP(stubbed, httptest.NewRequest(http.MethodGet, "/v1/users/stubbed", nil)))
	assert.Equal(t, http.StatusOK, stubbed.Code)
	assert.Contains(t, stubbed.Body.String(), "stubbed")

	unstubbed := httptest.NewRecorder()
	assert.True(t, dispatcher.ServeHTTP(unstubbed, httptest.NewRequest(http.MethodGet, "/v1/users/other", nil)))
	assert.Equal(t, http.StatusNotImplemented, unstubbed.Code)
	assert.Contains(t, unstubbed.Body.String(), "no matching stub")
}

func TestEngineDispatcherRejectsUnreadableRequestBody(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info: {title: Test API, version: 1.0.0}
paths: {}
`))
	require.NoError(t, err)
	router, err := openapi.NewRouter(doc)
	require.NoError(t, err)
	dispatcher := NewDispatcher(newMockStubRegistry(), router, 5, false, false)

	bodyErr := errors.New("request body read failed")
	body := &failingReadCloser{contents: []byte("partial"), err: bodyErr}
	request := httptest.NewRequest(http.MethodPost, "/items", nil)
	request.Body = body
	recorder := httptest.NewRecorder()

	assert.True(t, dispatcher.ServeHTTP(recorder, request))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), bodyErr.Error())
	assert.True(t, body.closed)
	restored, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("partial"), restored)
}

type failingReadCloser struct {
	contents []byte
	err      error
	read     bool
	closed   bool
}

func (b *failingReadCloser) Read(target []byte) (int, error) {
	if b.read {
		return 0, io.EOF
	}
	b.read = true
	return copy(target, b.contents), b.err
}

func (b *failingReadCloser) Close() error {
	b.closed = true
	return nil
}
