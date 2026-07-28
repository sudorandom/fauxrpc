package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/openapi"
	"github.com/sudorandom/fauxrpc/private/openapi/generator"
)

type StubRegistry interface {
	AddStub(rule StubRule)
	FindMatch(req *NormalizedRequest) (*StubRule, bool)
	Clear()
	NumStubs() int
}

const maxRequestBodySize int64 = 10 << 20

type Dispatcher struct {
	stubRegistry StubRegistry
	router       *openapi.Router
	validator    *openapi.Validator
	walker       *generator.Walker
	maxDepth     int
	onlyStubs    bool
	logger       *fauxlog.Logger
}

func NewDispatcher(stubReg StubRegistry, router *openapi.Router, maxDepth int, staticSeed, onlyStubs bool, logger *fauxlog.Logger) *Dispatcher {
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
		logger:       logger,
	}
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, req *http.Request) bool {
	if d.router == nil {
		return false
	}

	// 1. Route matching (do not read request body before route match)
	routeMatch, err := d.router.Match(req)
	if err != nil {
		return false
	}

	// 2. Read request body bytes if present
	var bodyBytes []byte
	if req.Body != nil {
		originalBody := req.Body
		var err error
		bodyBytes, err = io.ReadAll(io.LimitReader(originalBody, maxRequestBodySize+1))
		_ = originalBody.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("400 Bad Request: failed to read request body: %v", err), http.StatusBadRequest)
			return true
		}
		if int64(len(bodyBytes)) > maxRequestBodySize {
			http.Error(w, fmt.Sprintf("413 Request Entity Too Large: request body exceeds %d bytes", maxRequestBodySize), http.StatusRequestEntityTooLarge)
			return true
		}
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

	startTime := time.Now()

	logRequest := func(status int, resHeaders map[string]string, resBody []byte) {
		if d.logger == nil {
			return
		}
		reqHeadersJSON, _ := json.Marshal(req.Header)
		resHeadersJSON, _ := json.Marshal(resHeaders)
		d.logger.Log(&fauxlog.LogEntry{
			ID:              uuid.New().String(),
			Timestamp:       startTime,
			Service:         "OpenAPI",
			Method:          fmt.Sprintf("%s %s", req.Method, req.URL.Path),
			ClientProtocol:  "HTTP",
			Status:          status,
			Duration:        time.Since(startTime),
			RequestHeaders:  reqHeadersJSON,
			ResponseHeaders: resHeadersJSON,
			RequestBody:     bodyBytes,
			ResponseBody:    resBody,
		})
	}

	// 3. Pre-flight Validation
	if err := d.validator.ValidateRequest(req, routeMatch); err != nil {
		errMsg := fmt.Sprintf("400 Bad Request: %v", err)
		http.Error(w, errMsg, http.StatusBadRequest)
		logRequest(http.StatusBadRequest, map[string]string{"Content-Type": "text/plain; charset=utf-8"}, []byte(errMsg))
		return true
	}

	// 4. Unified Stub matching
	matchedStub, found := d.stubRegistry.FindMatch(normReq)
	if found {
		resBody := d.writeResponse(w, matchedStub.Response.Status, matchedStub.Response.Headers, matchedStub.Response.Body)
		logRequest(matchedStub.Response.Status, matchedStub.Response.Headers, resBody)
		return true
	}
	if d.onlyStubs {
		errMsg := "501 Not Implemented: no matching stub"
		http.Error(w, errMsg, http.StatusNotImplemented)
		logRequest(http.StatusNotImplemented, map[string]string{"Content-Type": "text/plain; charset=utf-8"}, []byte(errMsg))
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
			errMsg := fmt.Sprintf("500 Internal Server Error: %v", err)
			http.Error(w, errMsg, http.StatusInternalServerError)
			logRequest(http.StatusInternalServerError, map[string]string{"Content-Type": "text/plain; charset=utf-8"}, []byte(errMsg))
			return true
		}
		resBody := d.writeResponse(w, status, headers, payload)
		logRequest(status, headers, resBody)
		return true
	}

	return false
}

func (d *Dispatcher) writeResponse(w http.ResponseWriter, status int, headers map[string]string, body any) []byte {
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
				return nil
			}
		}
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(status)
	if len(encodedBody) > 0 {
		_, _ = w.Write(encodedBody)
	}
	return encodedBody
}
