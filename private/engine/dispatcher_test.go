package engine

import (
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
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

	dispatcher := NewDispatcher(registry, router, 5)

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
	assert.Contains(t, rec2.Body.String(), "id")
}
