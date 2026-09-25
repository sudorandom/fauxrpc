package openapi

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
)

type ValidatorOption func(*Validator)

// WithAuthenticationFunc configures a custom AuthenticationFunc for OpenAPI request validation.
func WithAuthenticationFunc(fn openapi3filter.AuthenticationFunc) ValidatorOption {
	return func(v *Validator) {
		v.options.AuthenticationFunc = fn
	}
}

type Validator struct {
	options *openapi3filter.Options
}

func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{
		options: &openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *Validator) ValidateRequest(req *http.Request, routeMatch *RouteMatch) error {
	input := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: routeMatch.PathParams,
		Route:      routeMatch.Route,
		Options:    v.options,
	}

	return openapi3filter.ValidateRequest(req.Context(), input)
}
