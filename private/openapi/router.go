package openapi

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

type Router struct {
	doc    *openapi3.T
	router routers.Router
}

type RouteMatch struct {
	Route       *routers.Route
	PathParams  map[string]string
	OperationID string
	Canonical   string
}

func NewRouter(doc *openapi3.T) (*Router, error) {
	r, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, err
	}
	return &Router{
		doc:    doc,
		router: r,
	}, nil
}

func (r *Router) Match(req *http.Request) (*RouteMatch, error) {
	route, pathParams, err := r.router.FindRoute(req)
	if err != nil {
		return nil, err
	}

	opID := ""
	if route.Operation != nil {
		opID = route.Operation.OperationID
	}

	return &RouteMatch{
		Route:       route,
		PathParams:  pathParams,
		OperationID: opID,
		Canonical:   route.PathItem.Ref, // Canonical path item ref or path pattern
	}, nil
}
