package engine

import (
	"net/http"
)

// NormalizedRequest standardizes requests across gRPC/Connect and HTTP/REST.
type NormalizedRequest struct {
	Service     string            `json:"service,omitempty"`
	Method      string            `json:"method,omitempty"`
	OperationID string            `json:"operationId,omitempty"`
	Path        string            `json:"path,omitempty"`
	HTTPMethod  string            `json:"httpMethod,omitempty"`
	PathParams  map[string]string `json:"pathParams,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
	Headers     http.Header       `json:"headers,omitempty"`
	Body        []byte            `json:"body,omitempty"`
}

// TargetRef identifies the target endpoint for stub matching.
type TargetRef struct {
	Service     string `json:"service,omitempty" yaml:"service,omitempty"`
	Method      string `json:"method,omitempty" yaml:"method,omitempty"`
	OperationID string `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Path        string `json:"path,omitempty" yaml:"path,omitempty"`
	HTTPMethod  string `json:"httpMethod,omitempty" yaml:"httpMethod,omitempty"`
}

// StubRule represents a unified stub rule for gRPC or REST.
type StubRule struct {
	ID          string       `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string       `json:"name,omitempty" yaml:"name,omitempty"`
	Target      TargetRef    `json:"target" yaml:"target"`
	Match       StubMatch    `json:"match,omitempty" yaml:"match,omitempty"`
	Response    StubResponse `json:"response" yaml:"response"`
	Priority    int          `json:"priority,omitempty" yaml:"priority,omitempty"`
	Specificity int          `json:"-" yaml:"-"`
}

// StubMatch holds predicates to evaluate against NormalizedRequest.
type StubMatch struct {
	PathParams     map[string]string `json:"pathParams,omitempty" yaml:"pathParams,omitempty"`
	QueryParams    map[string]string `json:"queryParams,omitempty" yaml:"queryParams,omitempty"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	BodyExpression string            `json:"bodyExpression,omitempty" yaml:"bodyExpression,omitempty"`
}

// StubResponse holds response details.
type StubResponse struct {
	Status  int               `json:"status" yaml:"status"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    any               `json:"body,omitempty" yaml:"body,omitempty"`
}
