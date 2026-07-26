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
	onlyStubs    bool
}

func NewDispatcher(stubReg StubRegistry, router *openapi.Router, maxDepth int, staticSeed, onlyStubs bool) *Dispatcher {
	if maxDepth <= 0 {
		maxDepth = 5
	}
	return &Dispatcher{
		stubRegistry: stubReg,
		router:       router,
		validator:    openapi.NewValidator(),
		walker:       generator.NewWalker(staticSeed),
		maxDepth:     maxDepth,
		onlyStubs:    onlyStubs,
	}
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, req *http.Request) bool {
	if d.router == nil {
		return false
	}

	// 1. Read request body bytes if present
	var bodyBytes []byte
	if req.Body != nil {
		originalBody := req.Body
		var err error
		bodyBytes, err = io.ReadAll(originalBody)
		_ = originalBody.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("400 Bad Request: failed to read request body: %v", err), http.StatusBadRequest)
			return true
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
	if d.onlyStubs {
		http.Error(w, "501 Not Implemented: no matching stub", http.StatusNotImplemented)
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

	var encodedBody []byte
	if body != nil {
		switch value := body.(type) {
		case []byte:
			encodedBody = value
		case string:
			encodedBody = []byte(value)
		default:
			var err error
			encodedBody, err = json.Marshal(value)
			if err != nil {
				http.Error(w, fmt.Sprintf("500 Internal Server Error: failed to encode response: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(status)
	if len(encodedBody) > 0 {
		_, _ = w.Write(encodedBody)
	}
}
