package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/stubs"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
)

type mockElizaImpl struct {
	elizav1connect.UnimplementedElizaServiceHandler
	sayFunc       func(context.Context, *connect.Request[elizav1.SayRequest]) (*connect.Response[elizav1.SayResponse], error)
	introduceFunc func(context.Context, *connect.Request[elizav1.IntroduceRequest], *connect.ServerStream[elizav1.IntroduceResponse]) error
}

func (m *mockElizaImpl) Say(ctx context.Context, req *connect.Request[elizav1.SayRequest]) (*connect.Response[elizav1.SayResponse], error) {
	if m.sayFunc != nil {
		return m.sayFunc(ctx, req)
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("unimplemented"))
}

func (m *mockElizaImpl) Introduce(ctx context.Context, req *connect.Request[elizav1.IntroduceRequest], stream *connect.ServerStream[elizav1.IntroduceResponse]) error {
	if m.introduceFunc != nil {
		return m.introduceFunc(ctx, req, stream)
	}
	return connect.NewError(connect.CodeUnimplemented, errors.New("unimplemented"))
}

func setupMockUpstream(t *testing.T, path string, handler http.Handler) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewUnstartedServer(mux)
	srv.EnableHTTP2 = true
	srv.StartTLS()
	return srv
}

func setupFauxRPCProxyServer(t *testing.T, upstreamURL string, recordDir string) (*server, *httptest.Server, elizav1connect.ElizaServiceClient) {
	reg := mustNewRegistry()
	fd := elizav1.File_connectrpc_eliza_v1_eliza_proto
	err := reg.RegisterFile(fd)
	require.NoError(t, err)

	opts := ServerOpts{
		Version:       "test-version",
		Addr:          "127.0.0.1:0",
		ProxyTo:       upstreamURL,
		RecordDir:     recordDir,
		UseReflection: false,
	}

	srv, err := NewServer(opts)
	require.NoError(t, err)
	srv.ServiceRegistry = reg

	err = srv.rebuildHandlers()
	require.NoError(t, err)

	mux, err := srv.Handler()
	require.NoError(t, err)
	ts := httptest.NewServer(mux)

	tr := &http.Transport{}
	tr.Protocols = new(http.Protocols)
	tr.Protocols.SetUnencryptedHTTP2(true)
	hc := &http.Client{
		Transport: tr,
	}
	client := elizav1connect.NewElizaServiceClient(
		hc,
		ts.URL,
	)
	return srv, ts, client
}

func TestProxyIngestion(t *testing.T) {
	recordDir := t.TempDir()

	sayCalled := false
	path, handler := elizav1connect.NewElizaServiceHandler(&mockElizaImpl{
		sayFunc: func(ctx context.Context, req *connect.Request[elizav1.SayRequest]) (*connect.Response[elizav1.SayResponse], error) {
			sayCalled = true
			assert.Equal(t, "Hello upstream", req.Msg.Sentence)
			return connect.NewResponse(&elizav1.SayResponse{
				Sentence: "Hello from upstream response",
			}), nil
		},
	})
	upstream := setupMockUpstream(t, path, handler)
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, recordDir)
	defer fauxrpc.Close()

	resp, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{
		Sentence: "Hello upstream",
	}))
	require.NoError(t, err)
	assert.True(t, sayCalled)
	assert.Equal(t, "Hello from upstream response", resp.Msg.Sentence)
	assert.Equal(t, "proxy", resp.Header().Get("x-fauxrpc-source"))

	// Verify recorded stubs file in its structured location
	recordedFilePath := filepath.Join(recordDir, "connectrpc.eliza.v1.ElizaService/Say.json")
	require.Eventually(t, func() bool {
		_, err := os.Stat(recordedFilePath)
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	recordedData, err := os.ReadFile(recordedFilePath)
	require.NoError(t, err)

	var stubFile stubs.StubFile
	err = json.Unmarshal(recordedData, &stubFile)
	require.NoError(t, err)

	require.Len(t, stubFile.Stubs, 1)
	stub := stubFile.Stubs[0]
	assert.Equal(t, "connectrpc.eliza.v1.ElizaService/Say", stub.Target)
	assert.Equal(t, `req.sentence == "Hello upstream"`, stub.ActiveIf)

	contentMap, ok := stub.Content.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello from upstream response", contentMap["sentence"])
}

func TestProxyUnimplementedFallback(t *testing.T) {
	path, handler := elizav1connect.NewElizaServiceHandler(&mockElizaImpl{})
	upstream := setupMockUpstream(t, path, handler)
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, "")
	defer fauxrpc.Close()

	resp, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{
		Sentence: "Hello upstream fallback",
	}))
	require.NoError(t, err)

	assert.Equal(t, "fake", resp.Header().Get("x-fauxrpc-source"))
	assert.Empty(t, resp.Header().Get("x-fauxrpc-mock-ids"))
	assert.NotEmpty(t, resp.Msg.Sentence)
}

func TestProxyUnimplementedFallbackWithStub(t *testing.T) {
	path, handler := elizav1connect.NewElizaServiceHandler(&mockElizaImpl{})
	upstream := setupMockUpstream(t, path, handler)
	defer upstream.Close()

	srv, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, "")
	defer fauxrpc.Close()

	stubKey := stubs.StubKey{
		Name: "connectrpc.eliza.v1.ElizaService.Say",
		ID:   "my-special-stub-id",
	}
	srv.AddStub(stubs.StubEntry{
		Key: stubKey,
		Message: &elizav1.SayResponse{
			Sentence: "Fallback stub sentence",
		},
	})

	resp, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{
		Sentence: "Hello upstream fallback stub",
	}))
	require.NoError(t, err)

	assert.Equal(t, "stub", resp.Header().Get("x-fauxrpc-source"))
	assert.Equal(t, "my-special-stub-id", resp.Header().Get("x-fauxrpc-mock-ids"))
	assert.Equal(t, "Fallback stub sentence", resp.Msg.Sentence)
}

func TestProxyIntegrationUnimplementedAndImplemented(t *testing.T) {
	sayCalled := false
	path, handler := elizav1connect.NewElizaServiceHandler(&mockElizaImpl{
		sayFunc: func(ctx context.Context, req *connect.Request[elizav1.SayRequest]) (*connect.Response[elizav1.SayResponse], error) {
			sayCalled = true
			return connect.NewResponse(&elizav1.SayResponse{
				Sentence: "Real response from implemented upstream endpoint",
			}), nil
		},
	})
	upstream := setupMockUpstream(t, path, handler)
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, "")
	defer fauxrpc.Close()

	// A. Implemented endpoint
	{
		resp, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{
			Sentence: "Hello upstream",
		}))
		require.NoError(t, err)
		assert.True(t, sayCalled)
		assert.Equal(t, "proxy", resp.Header().Get("x-fauxrpc-source"))
		assert.Equal(t, "Real response from implemented upstream endpoint", resp.Msg.Sentence)
	}

	// B. Unimplemented endpoint (Introduce) - fallback
	{
		stream, err := client.Introduce(context.Background(), connect.NewRequest(&elizav1.IntroduceRequest{
			Name: "Hello fallback",
		}))
		require.NoError(t, err)

		assert.Equal(t, "fake", stream.ResponseHeader().Get("x-fauxrpc-source"))
		assert.True(t, stream.Receive())
		assert.NotEmpty(t, stream.Msg().Sentence)
	}

	// C. Unimplemented bidi streaming endpoint (Converse) - fallback
	{
		stream := client.Converse(context.Background())
		err := stream.Send(&elizav1.ConverseRequest{
			Sentence: "Hello bidi fallback",
		})
		require.NoError(t, err)
		err = stream.CloseRequest()
		require.NoError(t, err)

		resp, err := stream.Receive()
		require.NoError(t, err)
		assert.Equal(t, "fake", stream.ResponseHeader().Get("x-fauxrpc-source"))
		assert.NotEmpty(t, resp.Sentence)
	}
}
