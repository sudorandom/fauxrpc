package server

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/proto"
)

type mockServer struct {
	registry.ServiceRegistry
	stubs.StubDatabase
	logger *fauxlog.Logger
}

func (m *mockServer) GetStats() *metrics.Stats          { return nil }
func (m *mockServer) IncrementTotalRequests()           {}
func (m *mockServer) IncrementErrors()                  {}
func (m *mockServer) GetLogger() *fauxlog.Logger        { return m.logger }

func TestHandler_Logging_Streaming(t *testing.T) {
	// Setup
	logger := fauxlog.NewLogger()
	logCh, unsubscribe := logger.Subscribe()
	defer unsubscribe()

	s := &mockServer{
		ServiceRegistry: mustNewRegistry(),
		StubDatabase:    stubs.NewStubDatabase(),
		logger:          logger,
	}

	validator, err := protovalidate.New()
	require.NoError(t, err)

	faker := fauxrpc.NewFauxFaker()

	// Eliza Service
	file := elizav1.File_connectrpc_eliza_v1_eliza_proto
	service := file.Services().ByName("ElizaService")
	require.NotNil(t, service)

	handler := NewHandler(service, faker, validator, s, logger)

	// Test Client Streaming (Converse is Bidi, so it counts as client streaming)
	converseMethod := service.Methods().ByName("Converse")
	require.NotNil(t, converseMethod)
	require.True(t, converseMethod.IsStreamingClient())

	// Create a pipe to simulate streaming body
	pr, pw := io.Pipe()

	req := httptest.NewRequest("POST", "/connectrpc.eliza.v1.ElizaService/Converse", pr)
	req.Header.Set("Content-Type", "application/grpc")

	w := httptest.NewRecorder()

	// Start handler in goroutine
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(w, req)
		close(done)
	}()

	// Write some messages to the pipe
	msg1 := &elizav1.ConverseRequest{Sentence: "Hello"}
	writeMsg(t, pw, msg1)

	msg2 := &elizav1.ConverseRequest{Sentence: "World"}
	writeMsg(t, pw, msg2)

	pw.Close()

	<-done

	// Verify logs
	select {
	case entry := <-logCh:
		assert.Equal(t, "connectrpc.eliza.v1.ElizaService", entry.Service)
		assert.Equal(t, "Converse", entry.Method)

		assert.Len(t, entry.RequestFrames, 2)

		var req1 map[string]interface{}
		err := json.Unmarshal(entry.RequestFrames[0], &req1)
		require.NoError(t, err)
		assert.Equal(t, "Hello", req1["sentence"])

		var req2 map[string]interface{}
		err = json.Unmarshal(entry.RequestFrames[1], &req2)
		require.NoError(t, err)
		assert.Equal(t, "World", req2["sentence"])

	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log entry")
	}
}

func mustNewRegistry() registry.ServiceRegistry {
	r, err := registry.NewServiceRegistry()
	if err != nil {
		panic(err)
	}
	return r
}

func writeMsg(t *testing.T, w io.Writer, msg proto.Message) {
	b, err := proto.Marshal(msg)
	require.NoError(t, err)

	// Prefix: 0 (not compressed) + 4 bytes length (big endian)
	prefix := make([]byte, 5)
	length := len(b)
	prefix[1] = byte(length >> 24)
	prefix[2] = byte(length >> 16)
	prefix[3] = byte(length >> 8)
	prefix[4] = byte(length)

	_, err = w.Write(prefix)
	require.NoError(t, err)
	_, err = w.Write(b)
	require.NoError(t, err)
}
