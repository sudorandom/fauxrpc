package server

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/elizav1connect"
)

// TestHandler_GzipRequest_UnaryResponse sends a gzip-compressed request to
// the unary Say method and verifies the server decodes it correctly.
func TestHandler_GzipRequest_UnaryResponse(t *testing.T) {
	_, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)

	for _, opts := range [][]connecthttp.Option{
		{connecthttp.WithGRPC(), connecthttp.WithSendGzip()},
		{connecthttp.WithSendGzip()},
	} {
		client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, opts...)))
		resp, err := client.Say(context.Background(), &elizav1.SayRequest{Sentence: "hello from gzip"})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Sentence)
	}
}

// TestHandler_GzipRequest_StreamingResponse sends gzip-compressed streaming
// frames and verifies the server decodes each one.
func TestHandler_GzipRequest_StreamingResponse(t *testing.T) {
	srv, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	logCh, unsubscribe := srv.logger.Subscribe()
	defer unsubscribe()

	client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC(), connecthttp.WithSendGzip())))
	stream, err := client.Converse(context.Background())
	require.NoError(t, err)
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "ping gzip 1"}))
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "ping gzip 2"}))
	require.NoError(t, stream.CloseSend())

	resp, err := stream.Receive()
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Sentence)
	for {
		if _, err := stream.Receive(); err != nil {
			require.True(t, errors.Is(err, io.EOF), "expected clean end of stream, got %v", err)
			break
		}
	}

	select {
	case entry := <-logCh:
		assert.Len(t, entry.RequestFrames, 2)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for log entry")
	}
}
