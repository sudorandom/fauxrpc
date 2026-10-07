package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/elizav1connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func mustNewRegistry() registry.ServiceRegistry {
	r, err := registry.NewServiceRegistry()
	if err != nil {
		panic(err)
	}
	return r
}

// setupTestServer builds a full FauxRPC server hosting the given files and
// serves it over HTTP/1.1 and unencrypted HTTP/2, returning an HTTP client
// that speaks both.
func setupTestServer(tb testing.TB, opts ServerOpts, files ...protoreflect.FileDescriptor) (*server, *httptest.Server, *http.Client) {
	tb.Helper()
	reg := mustNewRegistry()
	for _, fd := range files {
		require.NoError(tb, reg.RegisterFile(fd))
	}
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:0"
	}
	srv, err := NewServer(opts)
	require.NoError(tb, err)
	srv.ServiceRegistry = reg

	mux, err := srv.Handler()
	require.NoError(tb, err)
	ts := httptest.NewUnstartedServer(mux)
	ts.Config.Protocols = new(http.Protocols)
	ts.Config.Protocols.SetHTTP1(true)
	ts.Config.Protocols.SetUnencryptedHTTP2(true)
	ts.Start()
	tb.Cleanup(ts.Close)

	tr := &http.Transport{}
	tr.Protocols = new(http.Protocols)
	tr.Protocols.SetUnencryptedHTTP2(true)
	return srv, ts, &http.Client{Transport: tr}
}

func TestHandler_Logging_Streaming(t *testing.T) {
	srv, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	logCh, unsubscribe := srv.logger.Subscribe()
	defer unsubscribe()

	client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC())))
	stream, err := client.Converse(context.Background())
	require.NoError(t, err)
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "Hello"}))
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "World"}))
	require.NoError(t, stream.CloseSend())

	// The fake bidi handler responds with a single message and then closes.
	_, err = stream.Receive()
	require.NoError(t, err)
	for {
		if _, err := stream.Receive(); err != nil {
			require.True(t, errors.Is(err, io.EOF), "expected clean end of stream, got %v", err)
			break
		}
	}

	select {
	case entry := <-logCh:
		assert.Equal(t, "connectrpc.eliza.v1.ElizaService", entry.Service)
		assert.Equal(t, "Converse", entry.Method)
		assert.Equal(t, "gRPC", entry.ClientProtocol)
		assert.Equal(t, 0, entry.Status)

		require.Len(t, entry.RequestFrames, 2)
		var req1 map[string]any
		require.NoError(t, json.Unmarshal(entry.RequestFrames[0], &req1))
		assert.Equal(t, "Hello", req1["sentence"])
		var req2 map[string]any
		require.NoError(t, json.Unmarshal(entry.RequestFrames[1], &req2))
		assert.Equal(t, "World", req2["sentence"])

		require.NotEmpty(t, entry.ResponseFrames)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for log entry")
	}
}

func TestHandler_StaticSeedProducesStableProtobufResponses(t *testing.T) {
	_, ts, httpClient := setupTestServer(t,
		ServerOpts{StaticSeed: true},
		elizav1.File_connectrpc_eliza_v1_eliza_proto)

	client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC())))
	request := func() *elizav1.SayResponse {
		resp, err := client.Say(context.Background(), &elizav1.SayRequest{Sentence: "hello"})
		require.NoError(t, err)
		return resp
	}

	first, second := request(), request()
	assert.True(t, proto.Equal(first, second), "expected identical responses, got %v and %v", first, second)
}

// TestHandler_AllProtocols exercises the same unary RPC over each protocol
// connecthttp serves.
func TestHandler_AllProtocols(t *testing.T) {
	_, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)

	protocols := map[string][]connecthttp.Option{
		"connect":  nil,
		"grpc":     {connecthttp.WithGRPC()},
		"grpcweb":  {connecthttp.WithGRPCWeb()},
		"jsoncall": {connecthttp.WithProtoJSON()},
	}
	for name, opts := range protocols {
		t.Run(name, func(t *testing.T) {
			client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, opts...)))
			callCtx, callInfo := connect.NewClientContext(context.Background())
			resp, err := client.Say(callCtx, &elizav1.SayRequest{Sentence: "hello"})
			require.NoError(t, err)
			assert.NotEmpty(t, resp.Sentence)
			assert.Equal(t, "fake", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
		})
	}
}

// TestHandler_UnknownMethod verifies RPC-shaped requests for a method the
// service does not define get a proper RPC error.
func TestHandler_UnknownMethod(t *testing.T) {
	_, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)

	client := connect.NewClient(
		connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC()),
	)
	var resp elizav1.SayResponse
	err := client.CallUnary(context.Background(), connect.Spec{
		Procedure:  "/connectrpc.eliza.v1.ElizaService/NoSuchMethod",
		StreamType: connect.StreamTypeUnary,
	}, &elizav1.SayRequest{Sentence: "hi"}, &resp)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}
