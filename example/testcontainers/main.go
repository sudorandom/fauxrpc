package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	registryv1 "github.com/sudorandom/fauxrpc/proto/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/stubs/v1/stubsv1connect"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
)

func main() {
	ctx := context.Background()
	endpoint, err := setupFauxEliza(ctx)
	if err != nil {
		log.Fatalf("unable to set up faux eliza: %s", err)
	}

	{
		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, endpoint)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{
			Sentence: "testing!",
		}))
		if err != nil {
			log.Fatalf("unable to call eliza.Say: %s", err)
		}
		fmt.Println(resp.Msg)
	}

	{
		stubclient := stubsv1connect.NewStubsServiceClient(http.DefaultClient, endpoint)
		if _, err := stubclient.AddStubs(ctx, connect.NewRequest(&stubsv1.AddStubsRequest{
			Stubs: []*stubsv1.Stub{
				{
					Ref: &stubsv1.StubRef{
						Id:     "1234",
						Target: "connectrpc.eliza.v1.ElizaService/SayResponse",
						// Target: "connectrpc.eliza.v1.SayResponse",
					},
					Content: &stubsv1.Stub_Json{
						Json: `{"sentence": "I am setting this text!"}`,
					},
				},
			},
		})); err != nil {
			log.Fatalf("unable to set up stubs: %s", err)
		}

		elizaClient := elizav1connect.NewElizaServiceClient(http.DefaultClient, endpoint)
		resp, err := elizaClient.Say(ctx, connect.NewRequest(&elizav1.SayRequest{
			Sentence: "testing!",
		}))
		if err != nil {
			log.Fatalf("unable to call eliza.Say: %s", err)
		}
		fmt.Println(resp.Msg)
	}
}

func setupFauxEliza(ctx context.Context) (string, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.io/sudorandom/fauxrpc:latest",
			ExposedPorts: []string{"6660/tcp"},
			WaitingFor:   wait.ForLog("Server started"),
			Cmd:          []string{"run", "--empty", "--addr=:6660"},
		},
		Started: true,
	})
	if err != nil {
		return "", err
	}

	endpoint, err := container.PortEndpoint(ctx, "6660/tcp", "http")
	if err != nil {
		return "", err
	}

	client := registryv1connect.NewRegistryServiceClient(http.DefaultClient, endpoint)
	if _, err := client.AddDescriptors(ctx, connect.NewRequest(&registryv1.AddDescriptorsRequest{
		Descriptors: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				protodesc.ToFileDescriptorProto(elizav1.File_connectrpc_eliza_v1_eliza_proto),
			},
		},
	})); err != nil {
		return "", err
	}
	return endpoint, nil
}
