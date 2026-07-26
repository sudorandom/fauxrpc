package engine

import (
	"net/http"
)

// NormalizedResponse represents a response returned by stubs or generator.
type NormalizedResponse struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers,omitempty"`
	Body    any         `json:"body,omitempty"`
	RawBody []byte      `json:"rawBody,omitempty"`
}
