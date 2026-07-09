package server

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

// writeMsgGzip writes a gzip-compressed gRPC length-prefixed frame (flag=1).
func writeMsgGzip(t *testing.T, w io.Writer, msg proto.Message) {
	t.Helper()
	b, err := proto.Marshal(msg)
	require.NoError(t, err)

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	_, err = gz.Write(b)
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	payload := compressed.Bytes()
	prefix := make([]byte, 5)
	prefix[0] = 1 // compressed
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(payload)))
	_, err = w.Write(prefix)
	require.NoError(t, err)
	_, err = w.Write(payload)
	require.NoError(t, err)
}

// readGzipFrame reads one gRPC length-prefixed frame and decompresses it if flag=1.
func readGzipFrame(t *testing.T, r io.Reader) ([]byte, bool) {
	t.Helper()
	prefix := make([]byte, 5)
	_, err := io.ReadFull(r, prefix)
	if err == io.EOF {
		return nil, false
	}
	require.NoError(t, err)

	isCompressed := prefix[0] == 1
	size := binary.BigEndian.Uint32(prefix[1:])
	if size == 0 {
		return nil, true
	}

	payload := make([]byte, size)
	_, err = io.ReadFull(r, payload)
	require.NoError(t, err)

	if !isCompressed {
		return payload, true
	}

	gr, err := gzip.NewReader(bytes.NewReader(payload))
	require.NoError(t, err)
	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	return decompressed, true
}

// newElizaHandler is a test helper to build a handler for the ElizaService.
func newElizaHandler(t *testing.T) http.Handler {
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

// TestHandler_GzipRequest_UnaryResponse sends a gzip-compressed request to the
// unary Say method and verifies the handler decodes it correctly and returns a
// valid (uncompressed) gRPC response.
func TestHandler_GzipRequest_UnaryResponse(t *testing.T) {
	handler := newElizaHandler(t)

	// Build a gzip-compressed SayRequest body.
	var body bytes.Buffer
	writeMsgGzip(t, &body, &elizav1.SayRequest{Sentence: "hello from gzip"})

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", &body)
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-encoding", "gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Response should carry grpc-encoding: gzip (mirrored from request).
	assert.Equal(t, "gzip", res.Header.Get("grpc-encoding"))

	// The response body should be a valid, decompressible gRPC frame.
	respBytes, ok := readGzipFrame(t, res.Body)
	require.True(t, ok)

	var respMsg elizav1.SayResponse
	require.NoError(t, proto.Unmarshal(respBytes, &respMsg))
	// FauxRPC generates a non-nil response; just check it parsed without error.

	// gRPC status trailer should be OK (0).
	// In httptest.ResponseRecorder, grpc trailers land in res.Trailer.
	grpcStatus := res.Trailer.Get("Grpc-Status")
	if grpcStatus == "" {
		grpcStatus = res.Header.Get("Grpc-Status")
	}
	assert.Equal(t, "0", grpcStatus)
}

// TestHandler_UncompressedRequest_UncompressedResponse verifies the baseline
// (no grpc-encoding header) still works correctly after the changes.
func TestHandler_UncompressedRequest_UncompressedResponse(t *testing.T) {
	handler := newElizaHandler(t)

	var body bytes.Buffer
	writeMsg(t, &body, &elizav1.SayRequest{Sentence: "hello plain"})

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say", &body)
	req.Header.Set("Content-Type", "application/grpc")
	// No grpc-encoding header.

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Empty(t, res.Header.Get("grpc-encoding"), "no grpc-encoding header expected")
	// gRPC status trailer should be OK (0).
	grpcStatus := res.Trailer.Get("Grpc-Status")
	if grpcStatus == "" {
		grpcStatus = res.Header.Get("Grpc-Status")
	}
	assert.Equal(t, "0", grpcStatus)
}

// TestHandler_GzipRequest_StreamingResponse sends a gzip-compressed streaming
// (bidi) request and verifies the handler decodes each frame correctly and
// responds with gzip-compressed frames.
func TestHandler_GzipRequest_StreamingResponse(t *testing.T) {
	file := elizav1.File_connectrpc_eliza_v1_eliza_proto
	service := file.Services().ByName("ElizaService")
	require.NotNil(t, service)

	validator, err := protovalidate.New()
	require.NoError(t, err)

	logger := fauxlog.NewLogger()
	logCh, unsubscribe := logger.Subscribe()
	defer unsubscribe()

	s := &mockServer{
		ServiceRegistry: mustNewRegistry(),
		StubDatabase:    stubs.NewStubDatabase(),
		logger:          logger,
	}

	handler := NewHandler(service, fauxrpc.NewFauxFaker(), validator, s, logger, 20)

	pr, pw := io.Pipe()
	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Converse", pr)
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-encoding", "gzip")

	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(w, req)
		close(done)
	}()

	writeMsgGzip(t, pw, &elizav1.ConverseRequest{Sentence: "ping gzip 1"})
	writeMsgGzip(t, pw, &elizav1.ConverseRequest{Sentence: "ping gzip 2"})
	require.NoError(t, pw.Close())

	<-done

	res := w.Result()
	assert.Equal(t, "gzip", res.Header.Get("grpc-encoding"))

	// Should have received at least one compressed response frame.
	respBytes, ok := readGzipFrame(t, res.Body)
	require.True(t, ok)
	var respMsg elizav1.ConverseResponse
	require.NoError(t, proto.Unmarshal(respBytes, &respMsg))

	// Verify logs captured both gzip-decoded request frames.
	select {
	case entry := <-logCh:
		assert.Len(t, entry.RequestFrames, 2)
	default:
		t.Fatal("expected log entry")
	}
}

// TestHandler_GzipRequest_InvalidPayload sends a frame with flag=1 but
// invalid gzip bytes and expects the handler to return a gRPC error status.
func TestHandler_GzipRequest_InvalidPayload(t *testing.T) {
	handler := newElizaHandler(t)

	// Craft a frame with compressed=1 but garbage payload.
	garbage := []byte("this is not gzip")
	prefix := make([]byte, 5)
	prefix[0] = 1
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(garbage)))
	body := append(prefix, garbage...)

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Say",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("grpc-encoding", "gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	// The handler must return a non-OK gRPC status — not a 200 with empty body.
	grpcStatus := res.Header.Get("Grpc-Status")
	assert.NotEqual(t, "0", grpcStatus, "expected non-OK grpc status for invalid gzip payload")
	assert.NotEmpty(t, grpcStatus)
	// Specifically, this should be NotFound (4) or Internal (13), not OK (0).
	assert.NotEqual(t, codes.OK.String(), grpcStatus)
}
