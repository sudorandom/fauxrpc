package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sudorandom/fauxrpc/private/openapi"
	"github.com/sudorandom/fauxrpc/private/openapi/generator"
)

type StubRegistry interface {
	AddStub(rule StubRule)
	FindMatch(req *NormalizedRequest) (*StubRule, bool)
	Clear()
	NumStubs() int
}

type Dispatcher struct {
	stubRegistry StubRegistry
	router       *openapi.Router
	validator    *openapi.Validator
	walker       *generator.Walker
	maxDepth     int
}

func NewDispatcher(stubReg StubRegistry, router *openapi.Router, maxDepth int, staticSeed bool) *Dispatcher {
	if maxDepth <= 0 {
		maxDepth = 5
	}
	return &Dispatcher{
		stubRegistry: stubReg,
		router:       router,
		validator:    openapi.NewValidator(),
		walker:       generator.NewWalker(staticSeed),
		maxDepth:     maxDepth,
	}
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, req *http.Request) bool {
	if d.router == nil {
		return false
	}

	// 1. Read request body bytes if present
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err == nil {
			_ = req.Body.Close()
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	// 2. Radix Route matching
	routeMatch, err := d.router.Match(req)
	if err != nil {
		return false
	}

	queryParams := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	normReq := &NormalizedRequest{
		OperationID: routeMatch.OperationID,
		Path:        routeMatch.Route.Path,
		HTTPMethod:  req.Method,
		PathParams:  routeMatch.PathParams,
		QueryParams: queryParams,
		Headers:     req.Header,
		Body:        bodyBytes,
	}

	// 3. Pre-flight Validation
	if err := d.validator.ValidateRequest(req, routeMatch); err != nil {
		http.Error(w, fmt.Sprintf("400 Bad Request: %v", err), http.StatusBadRequest)
		return true
	}

	// 4. Unified Stub matching
	matchedStub, found := d.stubRegistry.FindMatch(normReq)
	if found {
		d.writeResponse(w, matchedStub.Response.Status, matchedStub.Response.Headers, matchedStub.Response.Body)
		return true
	}

	// 5. Auto-Generator Fallback
	if routeMatch.Route != nil && routeMatch.Route.Operation != nil {
		status, headers, payload, err := d.walker.GenerateFromOperation(
			req.Method,
			routeMatch.Route.Path,
			routeMatch.OperationID,
			routeMatch.Route.Operation,
			d.maxDepth,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
			return true
		}
		d.writeResponse(w, status, headers, payload)
		return true
	}

	return false
}

func (d *Dispatcher) writeResponse(w http.ResponseWriter, status int, headers map[string]string, body any) {
	if status == 0 {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(status)

	if body != nil {
		switch v := body.(type) {
		case []byte:
			_, _ = w.Write(v)
		case string:
			_, _ = w.Write([]byte(v))
		default:
			b, err := json.Marshal(v)
			if err == nil {
				_, _ = w.Write(b)
			}
		}
	}
}
