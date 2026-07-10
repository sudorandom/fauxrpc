package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/proto"
)

// --- Unit tests for clientAcceptsGzip ---

func TestClientAcceptsGzip(t *testing.T) {
	cases := []struct {
		name           string
		grpcEncoding   string
		grpcAcceptEnc  string
		expectAccepted bool
	}{
		{
			name:           "no headers — no gzip",
			expectAccepted: false,
		},
		{
			name:           "grpc-encoding: gzip only",
			grpcEncoding:   "gzip",
			expectAccepted: true,
		},
		{
			name:           "grpc-encoding: identity only",
			grpcEncoding:   "identity",
			expectAccepted: false,
		},
		{
			name:           "grpc-accept-encoding: gzip only",
			grpcAcceptEnc:  "gzip",
			expectAccepted: true,
		},
		{
			name:           "grpc-accept-encoding: identity,gzip",
			grpcAcceptEnc:  "identity,gzip",
			expectAccepted: true,
		},
		{
			name:           "grpc-accept-encoding: gzip,deflate",
			grpcAcceptEnc:  "gzip,deflate",
			expectAccepted: true,
		},
		{
			name:           "grpc-accept-encoding: identity only",
			grpcAcceptEnc:  "identity",
			expectAccepted: false,
		},
		{
			name:          "grpc-accept-encoding: identity — grpc-encoding: gzip",
			grpcEncoding:  "gzip",
			grpcAcceptEnc: "identity",
			// grpc-accept-encoding is present but doesn't list gzip;
			// grpc-encoding fallback should NOT be used when accept is explicit.
			expectAccepted: false,
		},
		{
			name:           "grpc-accept-encoding with whitespace around gzip",
			grpcAcceptEnc:  "identity, gzip , deflate",
			expectAccepted: true,
		},
		{
			name:           "grpc-accept-encoding: GZIP (case-insensitive)",
			grpcAcceptEnc:  "GZIP",
			expectAccepted: true,
		},
		{
			name:           "grpc-accept-encoding: Gzip (mixed case)",
			grpcAcceptEnc:  "Gzip",
			expectAccepted: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", nil)
			require.NoError(t, err)
			if tc.grpcEncoding != "" {
				req.Header.Set("grpc-encoding", tc.grpcEncoding)
			}
			if tc.grpcAcceptEnc != "" {
				req.Header.Set("grpc-accept-encoding", tc.grpcAcceptEnc)
			}
			assert.Equal(t, tc.expectAccepted, clientAcceptsGzip(req))
		})
	}
}

func TestResponseCompressionEncoding(t *testing.T) {
	cases := []struct {
		name           string
		grpcEncoding   string
		grpcAcceptEnc  string
		expectEncoding string
	}{
		{
			name: "no compression headers",
		},
		{
			name:           "mirror gzip request when accept is absent",
			grpcEncoding:   "gzip",
			expectEncoding: "gzip",
		},
		{
			name:           "mirror deflate request when accept is absent",
			grpcEncoding:   "deflate",
			expectEncoding: "deflate",
		},
		{
			name:           "prefer gzip over deflate when both are accepted",
			grpcAcceptEnc:  "deflate,gzip",
			expectEncoding: "gzip",
		},
		{
			name:           "use deflate when it is the only supported accepted encoding",
			grpcAcceptEnc:  "identity, deflate",
			expectEncoding: "deflate",
		},
		{
			name:           "accept header overrides request encoding",
			grpcEncoding:   "deflate",
			grpcAcceptEnc:  "identity",
			expectEncoding: "",
		},
		{
			name:           "case-insensitive deflate",
			grpcAcceptEnc:  "DEFLATE",
			expectEncoding: "deflate",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", nil)
			require.NoError(t, err)
			if tc.grpcEncoding != "" {
				req.Header.Set("grpc-encoding", tc.grpcEncoding)
			}
			if tc.grpcAcceptEnc != "" {
				req.Header.Set("grpc-accept-encoding", tc.grpcAcceptEnc)
			}
			assert.Equal(t, tc.expectEncoding, responseCompressionEncoding(req))
		})
	}
}

// --- Integration tests: handler respects grpc-accept-encoding ---

// newTestHandler builds a handler for ElizaService using the faux faker.
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	file := elizav1.File_connectrpc_eliza_v1_eliza_proto
	service := file.Services().ByName("ElizaService")
	require.NotNil(t, service)
	validator, err := protovalidate.New()
	require.NoError(t, err)
	logger := fauxlog.NewLogger()
	s := &mockServer{
		ServiceRegistry: mustNewRegistry(),
		StubDatabase:    stubs.NewStubDatabase(),
		logger:          logger,
	}
	return NewHandler(service, fauxrpc.NewFauxFaker(), validator, s, logger, 20)
}

// makeUnaryBody returns a plain (uncompressed) gRPC-framed SayRequest body.
func makeUnaryBody(t *testing.T) *bytes.Buffer {
	t.Helper()
	var body bytes.Buffer
	writeMsg(t, &body, &elizav1.SayRequest{Sentence: "hello"})
	return &body
}

// TestHandler_AcceptEncoding_GzipOnlyInAccept verifies that a client sending an
// uncompressed request but advertising grpc-accept-encoding: gzip receives a
// gzip-compressed response.
func TestHandler_AcceptEncoding_GzipOnlyInAccept(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", makeUnaryBody(t))
	req.Header.Set("Content-Type", "application/grpc")
	// Client sends uncompressed but can decode gzip.
	req.Header.Set("grpc-accept-encoding", "gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, "gzip", res.Header.Get("grpc-encoding"),
		"server should compress response when client advertises grpc-accept-encoding: gzip")

	// Verify the frame is actually decompressible.
	respBytes, ok := readGzipFrame(t, res.Body)
	require.True(t, ok)
	var msg elizav1.SayResponse
	require.NoError(t, proto.Unmarshal(respBytes, &msg))
}

// TestHandler_AcceptEncoding_MultipleEncodings verifies a comma-separated
// grpc-accept-encoding that includes gzip triggers compressed responses.
func TestHandler_AcceptEncoding_MultipleEncodings(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", makeUnaryBody(t))
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-accept-encoding", "identity,gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, "gzip", res.Header.Get("grpc-encoding"))
}

func TestHandler_AcceptEncoding_DeflateOnly(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", makeUnaryBody(t))
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-accept-encoding", "deflate")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, "deflate", res.Header.Get("grpc-encoding"))

	respBytes, ok := readDeflateFrame(t, res.Body)
	require.True(t, ok)
	var msg elizav1.SayResponse
	require.NoError(t, proto.Unmarshal(respBytes, &msg))
}

// TestHandler_AcceptEncoding_IdentityOnly verifies that a client advertising
// only identity (no gzip) receives an uncompressed response, even if the
// grpc-encoding on the request happened to be something else.
func TestHandler_AcceptEncoding_IdentityOnly(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", makeUnaryBody(t))
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-accept-encoding", "identity")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Empty(t, res.Header.Get("grpc-encoding"),
		"server must not compress response when client only accepts identity")
}

// TestHandler_AcceptEncoding_GzipRequestAndAccept verifies the happy-path where
// both grpc-encoding and grpc-accept-encoding are gzip.
func TestHandler_AcceptEncoding_GzipRequestAndAccept(t *testing.T) {
	handler := newTestHandler(t)

	var body bytes.Buffer
	writeMsgGzip(t, &body, &elizav1.SayRequest{Sentence: "hello compressed"})

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", &body)
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-encoding", "gzip")
	req.Header.Set("grpc-accept-encoding", "gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, "gzip", res.Header.Get("grpc-encoding"))

	respBytes, ok := readGzipFrame(t, res.Body)
	require.True(t, ok)
	var msg elizav1.SayResponse
	require.NoError(t, proto.Unmarshal(respBytes, &msg))
}

// TestHandler_AcceptEncoding_NoHeaders verifies the baseline: no encoding
// headers means no compression in either direction.
func TestHandler_AcceptEncoding_NoHeaders(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", makeUnaryBody(t))
	req.Header.Set("Content-Type", "application/grpc")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Empty(t, res.Header.Get("grpc-encoding"))
}
