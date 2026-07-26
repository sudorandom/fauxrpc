package openapi

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
)

type Validator struct {
	options *openapi3filter.Options
}

func NewValidator() *Validator {
	return &Validator{
		options: &openapi3filter.Options{
			MultiError: true,
		},
	}
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
