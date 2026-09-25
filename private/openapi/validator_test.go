package openapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorSecuritySchemes(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
info:
  title: Sample Secured API
  version: 1.0.0
security:
  - GlobalAuth: []
paths:
  /hello:
    get:
      summary: Say Hello
      operationId: sayHello
      security:
        - BearerAuth: []
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  message: { type: string }
components:
  securitySchemes:
    GlobalAuth:
      type: http
      scheme: basic
    BearerAuth:
      type: http
      scheme: bearer
`)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	router, err := NewRouter(doc)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	routeMatch, err := router.Match(req)
	require.NoError(t, err)

	t.Run("default validator uses NoopAuthenticationFunc", func(t *testing.T) {
		validator := NewValidator()
		err := validator.ValidateRequest(req, routeMatch)
		assert.NoError(t, err)
	})

	t.Run("custom authentication func that succeeds", func(t *testing.T) {
		called := false
		customAuth := func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
			called = true
			assert.Equal(t, "BearerAuth", input.SecuritySchemeName)
			return nil
		}
		validator := NewValidator(WithAuthenticationFunc(customAuth))
		err := validator.ValidateRequest(req, routeMatch)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("custom authentication func that rejects", func(t *testing.T) {
		authErr := errors.New("unauthorized token")
		customAuth := func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
			return authErr
		}
		validator := NewValidator(WithAuthenticationFunc(customAuth))
		err := validator.ValidateRequest(req, routeMatch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized token")
	})
}
