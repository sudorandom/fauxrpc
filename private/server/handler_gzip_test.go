package server

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_GzipRequest_UnaryResponse sends a gzip-compressed request to
// the unary Say method and verifies the server decodes it correctly.
func TestHandler_GzipRequest_UnaryResponse(t *testing.T) {
	_, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)

	for _, opts := range [][]connect.ClientOption{
		{connect.WithGRPC(), connect.WithSendGzip()},
		{connect.WithSendGzip()},
	} {
		client := elizav1connect.NewElizaServiceClient(httpClient, ts.URL, opts...)
		resp, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{Sentence: "hello from gzip"}))
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Msg.Sentence)
	}
}

// TestHandler_GzipRequest_StreamingResponse sends gzip-compressed streaming
// frames and verifies the server decodes each one.
func TestHandler_GzipRequest_StreamingResponse(t *testing.T) {
	srv, ts, httpClient := setupTestServer(t, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	logCh, unsubscribe := srv.logger.Subscribe()
	defer unsubscribe()

	client := elizav1connect.NewElizaServiceClient(httpClient, ts.URL, connect.WithGRPC(), connect.WithSendGzip())
	stream := client.Converse(context.Background())
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "ping gzip 1"}))
	require.NoError(t, stream.Send(&elizav1.ConverseRequest{Sentence: "ping gzip 2"}))
	require.NoError(t, stream.CloseRequest())

	resp, err := stream.Receive()
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Sentence)
	for {
		if _, err := stream.Receive(); err != nil {
			require.True(t, errors.Is(err, io.EOF), "expected clean end of stream, got %v", err)
			break
		}
	}
	require.NoError(t, stream.CloseResponse())

	select {
	case entry := <-logCh:
		assert.Len(t, entry.RequestFrames, 2)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for log entry")
	}
}
