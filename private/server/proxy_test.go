package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/elizav1connect"
	"github.com/sudorandom/fauxrpc/private/stubs"
)

type mockElizaImpl struct {
	elizav1connect.UnimplementedElizaServiceHandler
	sayFunc func(context.Context, *elizav1.SayRequest) (*elizav1.SayResponse, error)
}

func (m *mockElizaImpl) Say(ctx context.Context, req *elizav1.SayRequest) (*elizav1.SayResponse, error) {
	if m.sayFunc != nil {
		return m.sayFunc(ctx, req)
	}
	return nil, connect.NewError(connect.CodeUnimplemented, "unimplemented")
}

func setupMockUpstream(t *testing.T, impl elizav1connect.ElizaServiceHandler) *httptest.Server {
	mux := http.NewServeMux()
	srv := connect.NewServer()
	elizav1connect.RegisterElizaServiceHandler(srv, impl)
	connecthttp.Mount(mux, srv)
	ts := httptest.NewUnstartedServer(mux)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	return ts
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
	ts := httptest.NewUnstartedServer(mux)
	ts.Config.Protocols = new(http.Protocols)
	ts.Config.Protocols.SetHTTP1(true)
	ts.Config.Protocols.SetUnencryptedHTTP2(true)
	ts.Start()

	tr := &http.Transport{}
	tr.Protocols = new(http.Protocols)
	tr.Protocols.SetUnencryptedHTTP2(true)
	hc := &http.Client{
		Transport: tr,
	}
	client := elizav1connect.NewElizaServiceClient(
		connect.NewClient(connecthttp.NewTransport(hc, ts.URL)),
	)
	return srv, ts, client
}

func TestProxyIngestion(t *testing.T) {
	recordDir := t.TempDir()

	sayCalled := false
	upstream := setupMockUpstream(t, &mockElizaImpl{
		sayFunc: func(ctx context.Context, req *elizav1.SayRequest) (*elizav1.SayResponse, error) {
			sayCalled = true
			assert.Equal(t, "Hello upstream", req.Sentence)
			return &elizav1.SayResponse{
				Sentence: "Hello from upstream response",
			}, nil
		},
	})
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, recordDir)
	defer fauxrpc.Close()

	callCtx, callInfo := connect.NewClientContext(context.Background())
	resp, err := client.Say(callCtx, &elizav1.SayRequest{
		Sentence: "Hello upstream",
	})
	require.NoError(t, err)
	assert.True(t, sayCalled)
	assert.Equal(t, "Hello from upstream response", resp.Sentence)
	assert.Equal(t, "proxy", callInfo.ResponseHeader().Get("x-fauxrpc-source"))

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
	upstream := setupMockUpstream(t, &mockElizaImpl{})
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, "")
	defer fauxrpc.Close()

	callCtx, callInfo := connect.NewClientContext(context.Background())
	resp, err := client.Say(callCtx, &elizav1.SayRequest{
		Sentence: "Hello upstream fallback",
	})
	require.NoError(t, err)

	assert.Equal(t, "fake", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
	assert.Empty(t, callInfo.ResponseHeader().Get("x-fauxrpc-mock-ids"))
	assert.NotEmpty(t, resp.Sentence)
}

func TestProxyUnimplementedFallbackWithStub(t *testing.T) {
	upstream := setupMockUpstream(t, &mockElizaImpl{})
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

	callCtx, callInfo := connect.NewClientContext(context.Background())
	resp, err := client.Say(callCtx, &elizav1.SayRequest{
		Sentence: "Hello upstream fallback stub",
	})
	require.NoError(t, err)

	assert.Equal(t, "stub", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
	assert.Equal(t, "my-special-stub-id", callInfo.ResponseHeader().Get("x-fauxrpc-mock-ids"))
	assert.Equal(t, "Fallback stub sentence", resp.Sentence)
}

func TestProxyIntegrationUnimplementedAndImplemented(t *testing.T) {
	sayCalled := false
	upstream := setupMockUpstream(t, &mockElizaImpl{
		sayFunc: func(ctx context.Context, req *elizav1.SayRequest) (*elizav1.SayResponse, error) {
			sayCalled = true
			return &elizav1.SayResponse{
				Sentence: "Real response from implemented upstream endpoint",
			}, nil
		},
	})
	defer upstream.Close()

	_, fauxrpc, client := setupFauxRPCProxyServer(t, upstream.URL, "")
	defer fauxrpc.Close()

	// A. Implemented endpoint
	{
		callCtx, callInfo := connect.NewClientContext(context.Background())
		resp, err := client.Say(callCtx, &elizav1.SayRequest{
			Sentence: "Hello upstream",
		})
		require.NoError(t, err)
		assert.True(t, sayCalled)
		assert.Equal(t, "proxy", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
		assert.Equal(t, "Real response from implemented upstream endpoint", resp.Sentence)
	}

	// B. Unimplemented endpoint (Introduce) - fallback
	{
		callCtx, callInfo := connect.NewClientContext(context.Background())
		stream, err := client.Introduce(callCtx, &elizav1.IntroduceRequest{
			Name: "Hello fallback",
		})
		require.NoError(t, err)

		msg, err := stream.Receive()
		require.NoError(t, err)
		assert.Equal(t, "fake", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
		assert.NotEmpty(t, msg.Sentence)
	}

	// C. Unimplemented bidi streaming endpoint (Converse) - fallback
	{
		callCtx, callInfo := connect.NewClientContext(context.Background())
		stream, err := client.Converse(callCtx)
		require.NoError(t, err)
		err = stream.Send(&elizav1.ConverseRequest{
			Sentence: "Hello bidi fallback",
		})
		require.NoError(t, err)
		err = stream.CloseSend()
		require.NoError(t, err)

		resp, err := stream.Receive()
		require.NoError(t, err)
		assert.Equal(t, "fake", callInfo.ResponseHeader().Get("x-fauxrpc-source"))
		assert.NotEmpty(t, resp.Sentence)
	}
}
