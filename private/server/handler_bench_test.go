package server

import (
	"context"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/elizav1connect"
)

func BenchmarkHandler_Unary(b *testing.B) {
	_, ts, httpClient := setupTestServer(b, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC())))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := client.Say(context.Background(), &elizav1.SayRequest{Sentence: "Hello World"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandler_Streaming_Messages(b *testing.B) {
	_, ts, httpClient := setupTestServer(b, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	client := elizav1connect.NewElizaServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, ts.URL, connecthttp.WithGRPC())))

	stream, err := client.Converse(context.Background())
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := stream.Send(&elizav1.ConverseRequest{Sentence: "Hello World"}); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	require.NoError(b, stream.CloseSend())
	for {
		if _, err := stream.Receive(); err != nil {
			break
		}
	}
}
