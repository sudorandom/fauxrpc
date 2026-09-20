package server

import (
	"context"
	"testing"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
)

func BenchmarkHandler_Unary(b *testing.B) {
	_, ts, httpClient := setupTestServer(b, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	client := elizav1connect.NewElizaServiceClient(httpClient, ts.URL, connect.WithGRPC())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{Sentence: "Hello World"}))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandler_Streaming_Messages(b *testing.B) {
	_, ts, httpClient := setupTestServer(b, ServerOpts{}, elizav1.File_connectrpc_eliza_v1_eliza_proto)
	client := elizav1connect.NewElizaServiceClient(httpClient, ts.URL, connect.WithGRPC())

	stream := client.Converse(context.Background())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := stream.Send(&elizav1.ConverseRequest{Sentence: "Hello World"}); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	require.NoError(b, stream.CloseRequest())
	for {
		if _, err := stream.Receive(); err != nil {
			break
		}
	}
	require.NoError(b, stream.CloseResponse())
}
