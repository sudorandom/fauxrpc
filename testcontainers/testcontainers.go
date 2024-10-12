package fauxrpctestcontainers

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	registryv1 "github.com/sudorandom/fauxrpc/proto/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/stubs/v1/stubsv1connect"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type FauxRPCContainer struct {
	testcontainers.Container
}

func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*FauxRPCContainer, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"6660/tcp"},
			WaitingFor:   wait.ForLog("Server started"),
			Cmd:          []string{"run", "--empty", "--addr=:6660"},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return &FauxRPCContainer{
		Container: container,
	}, nil
}

func (c *FauxRPCContainer) MustBaseURL(ctx context.Context) string {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		panic(err)
	}
	return baseURL
}

func (c *FauxRPCContainer) BaseURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "6660/tcp", "http")
	if err != nil {
		return "", err
	}
	return endpoint, nil
}

func (c *FauxRPCContainer) MustRegistryClient(ctx context.Context) registryv1connect.RegistryServiceClient {
	return registryv1connect.NewRegistryServiceClient(http.DefaultClient, c.MustBaseURL(ctx))
}

func (c *FauxRPCContainer) RegistryClient(ctx context.Context) (registryv1connect.RegistryServiceClient, error) {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		return nil, err
	}
	return registryv1connect.NewRegistryServiceClient(http.DefaultClient, baseURL), nil
}

func (c *FauxRPCContainer) MustAddFileDescriptor(ctx context.Context, fd protoreflect.FileDescriptor) {
	if err := c.AddFileDescriptor(ctx, fd); err != nil {
		panic(err)
	}
}

func (c *FauxRPCContainer) AddFileDescriptor(ctx context.Context, fd protoreflect.FileDescriptor) error {
	client, err := c.RegistryClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.AddDescriptors(ctx, connect.NewRequest(&registryv1.AddDescriptorsRequest{
		Descriptors: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				protodesc.ToFileDescriptorProto(fd),
			},
		},
	}))
	return err
}

func (c *FauxRPCContainer) MustResetRegistry(ctx context.Context) {
	if err := c.ResetRegistry(ctx); err != nil {
		panic(err)
	}
}

func (c *FauxRPCContainer) ResetRegistry(ctx context.Context) error {
	client, err := c.RegistryClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.Reset(ctx, connect.NewRequest(&registryv1.ResetRequest{}))
	return err
}

func (c *FauxRPCContainer) MustStubsClient(ctx context.Context) stubsv1connect.StubsServiceClient {
	return stubsv1connect.NewStubsServiceClient(http.DefaultClient, c.MustBaseURL(ctx))
}

func (c *FauxRPCContainer) StubsClient(ctx context.Context) (stubsv1connect.StubsServiceClient, error) {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		return nil, err
	}
	return stubsv1connect.NewStubsServiceClient(http.DefaultClient, baseURL), nil
}

func (c *FauxRPCContainer) MustAddStub(ctx context.Context, target string, msg proto.Message) {
	if err := c.AddStub(ctx, target, msg); err != nil {
		panic(err)
	}
}

func (c *FauxRPCContainer) AddStub(ctx context.Context, target string, msg proto.Message) error {
	client, err := c.StubsClient(ctx)
	if err != nil {
		return err
	}
	msgpb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = client.AddStubs(ctx, connect.NewRequest(&stubsv1.AddStubsRequest{
		Stubs: []*stubsv1.Stub{
			{
				Ref: &stubsv1.StubRef{
					Id:     uuid.New().String(),
					Target: target,
				},
				Content: &stubsv1.Stub_Proto{Proto: msgpb},
			},
		},
	}))
	return err
}

func (c *FauxRPCContainer) MustResetStubs(ctx context.Context) {
	if err := c.ResetStubs(ctx); err != nil {
		panic(err)
	}
}

func (c *FauxRPCContainer) ResetStubs(ctx context.Context) error {
	client, err := c.StubsClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.RemoveAllStubs(ctx, connect.NewRequest(&stubsv1.RemoveAllStubsRequest{}))
	return err
}
