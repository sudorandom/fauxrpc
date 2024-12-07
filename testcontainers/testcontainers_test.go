package fauxrpctestcontainers_test

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	fauxrpctestcontainers "github.com/sudorandom/fauxrpc/testcontainers"
	"google.golang.org/protobuf/reflect/protoregistry"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
)

func TestContainersTest(t *testing.T) {
	ctx := context.Background()

	// Start fauxrpc container
	container, err := fauxrpctestcontainers.Run(ctx, "docker.io/sudorandom/fauxrpc:latest")
	if err != nil {
		t.Fatalf("unable to set up faux eliza: %s", err)
	}
	t.Cleanup(func() { container.Terminate(context.Background()) })

	baseURL := container.MustBaseURL(ctx)

	t.Run("register files", func(t *testing.T) {
		container.MustAddFiles(ctx, protoregistry.GlobalFiles)
		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{
			Sentence: "testing!",
		}))
		if err != nil {
			t.Fatalf("unable to call eliza.Say: %s", err)
		}
		if len(resp.Msg.Sentence) == 0 {
			t.Fatal("sentence should not be empty, but it was")
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)

	// Now we can call the service and generated data will be returned
	t.Run("using the default generated responses", func(t *testing.T) {
		container.MustAddFileDescriptor(ctx, elizav1.File_connectrpc_eliza_v1_eliza_proto)

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{
			Sentence: "testing!",
		}))
		if err != nil {
			t.Fatalf("unable to call eliza.Say: %s", err)
		}
		if len(resp.Msg.Sentence) == 0 {
			t.Fatal("sentence should not be empty, but it was")
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)

	// We can also register stubs, to set up specific scenarios
	t.Run("using stubs responses", func(t *testing.T) {
		container.MustAddFileDescriptor(ctx, elizav1.File_connectrpc_eliza_v1_eliza_proto)
		container.MustAddStub(ctx, "connectrpc.eliza.v1.ElizaService/Say", &elizav1.SayResponse{Sentence: "I am setting this text!"})

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{Sentence: "testing!"}))
		if err != nil {
			t.Fatalf("unable to call eliza.Say: %s", err)
		}
		expected := "I am setting this text!"
		if resp.Msg.Sentence != expected {
			t.Fatalf("stubbed sentence does not match! %s != %s", resp.Msg.Sentence, expected)
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)

	// We can also register stubs based on the message type
	t.Run("using stubs responses on type", func(t *testing.T) {
		container.MustAddFileDescriptor(ctx, elizav1.File_connectrpc_eliza_v1_eliza_proto)
		container.MustAddStub(ctx, "connectrpc.eliza.v1.SayResponse", &elizav1.SayResponse{
			Sentence: "I am setting this text!",
		})

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{
			Sentence: "testing!",
		}))
		if err != nil {
			t.Fatalf("unable to call eliza.Say: %s", err)
		}
		expected := "I am setting this text!"
		if resp.Msg.Sentence != expected {
			t.Fatalf("stubbed sentence does not match! %s != %s", resp.Msg.Sentence, expected)
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)

	// We can also register error stubs
	t.Run("using stubs responses on type", func(t *testing.T) {
		container.MustAddFileDescriptor(ctx, elizav1.File_connectrpc_eliza_v1_eliza_proto)
		container.MustAddStubError(ctx, "connectrpc.eliza.v1.ElizaService/Say", "server down", stubsv1.ErrorCode_ERROR_CODE_UNAVAILABLE)

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{Sentence: "testing!"}))
		if err == nil {
			t.Fatalf("error was expected! but we didn't get one! resp=%v", resp.Msg)
		}
		expected := "unavailable: server down"
		if err.Error() != expected {
			t.Fatalf("stubbed error message does not match! %s != %s", err.Error(), expected)
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)

	// We can also add protobuf schema from a file path
	t.Run("add schema from path", func(t *testing.T) {
		container.MustAddFromPath(ctx, "eliza.binpb")
		container.MustAddStubError(ctx, "connectrpc.eliza.v1.ElizaService/Say", "server down", stubsv1.ErrorCode_ERROR_CODE_UNAVAILABLE)

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, baseURL)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{Sentence: "testing!"}))
		if err == nil {
			t.Fatalf("error was expected! but we didn't get one! resp=%v", resp.Msg)
		}
		expected := "unavailable: server down"
		if err.Error() != expected {
			t.Fatalf("stubbed error message does not match! %s != %s", err.Error(), expected)
		}
	})

	container.MustResetStubs(ctx)
	container.MustResetRegistry(ctx)
}
