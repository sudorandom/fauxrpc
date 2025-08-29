package fauxrpctestcontainers

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	registryv1 "github.com/sudorandom/fauxrpc/private/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/private/gen/registry/v1/registryv1connect"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/gen/stubs/v1/stubsv1connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type FauxRPCContainer struct {
	testcontainers.Container
}

// Run will start the testcontainers. You can pass in your own image name. It's advisable to avoid using
// the latest tag but if you want to, you would use "docker.io/sudorandom/fauxrpc:latest"
// See here for all available versions: https://hub.docker.com/r/sudorandom/fauxrpc
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

// MustBaseURL returns the base URL for the FauxRPC service. A panic happens if anything fails.
func (c *FauxRPCContainer) MustBaseURL(ctx context.Context) string {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		panic(err)
	}
	return baseURL
}

// BaseURL returns the base URL for the FauxRPC service. This will include a "http://" prefix.
func (c *FauxRPCContainer) BaseURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "6660/tcp", "http")
	if err != nil {
		return "", err
	}
	return endpoint, nil
}

// MustRegistryClient returns a client for managing types in the FauxRPC registry. A panic happens if anything fails.
func (c *FauxRPCContainer) MustRegistryClient(ctx context.Context) registryv1connect.RegistryServiceClient {
	return registryv1connect.NewRegistryServiceClient(http.DefaultClient, c.MustBaseURL(ctx))
}

// MustRegistryClient returns a client for managing types in the FauxRPC registry.
func (c *FauxRPCContainer) RegistryClient(ctx context.Context) (registryv1connect.RegistryServiceClient, error) {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		return nil, err
	}
	return registryv1connect.NewRegistryServiceClient(http.DefaultClient, baseURL), nil
}

// MustAddFileDescriptor adds the given file descriptor to the FauxRPC registry. A panic happens if anything fails.
func (c *FauxRPCContainer) MustAddFileDescriptor(ctx context.Context, fd protoreflect.FileDescriptor) {
	if err := c.AddFileDescriptor(ctx, fd); err != nil {
		panic(err)
	}
}

// MustAddFileDescriptor adds the given file descriptor to the FauxRPC registry.
func (c *FauxRPCContainer) AddFileDescriptor(ctx context.Context, fd protoreflect.FileDescriptor) error {
	client, err := c.RegistryClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.AddDescriptors(ctx, connect.NewRequest(
		registryv1.AddDescriptorsRequest_builder{
			Descriptors: &descriptorpb.FileDescriptorSet{
				File: []*descriptorpb.FileDescriptorProto{
					protodesc.ToFileDescriptorProto(fd),
				},
			},
		}.Build()))
	return err
}

// MustAddFromPath adds the given protoregistry.Files to the FauxRPC registry. A panic happens if anything fails.
func (c *FauxRPCContainer) MustAddFromPath(ctx context.Context, path string) {
	if err := c.AddFromPath(ctx, path); err != nil {
		panic(err)
	}
}

// AddFromPath adds the given protoregistry.Files to the FauxRPC registry.
func (c *FauxRPCContainer) AddFromPath(ctx context.Context, path string) error {
	files := new(protoregistry.Files)
	if err := registry.AddServicesFromPath(files, path); err != nil {
		return err
	}

	return c.AddFiles(ctx, files)
}

// MustAddFiles adds the given protoregistry.Files to the FauxRPC registry. A panic happens if anything fails.
func (c *FauxRPCContainer) MustAddFiles(ctx context.Context, files *protoregistry.Files) {
	if err := c.AddFiles(ctx, files); err != nil {
		panic(err)
	}
}

// AddFiles adds the given protoregistry.Files to the FauxRPC registry.
func (c *FauxRPCContainer) AddFiles(ctx context.Context, files *protoregistry.Files) error {
	client, err := c.RegistryClient(ctx)
	if err != nil {
		return err
	}
	fds := &descriptorpb.FileDescriptorSet{}
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fd))
		return true
	})

	_, err = client.AddDescriptors(ctx, connect.NewRequest(
		registryv1.AddDescriptorsRequest_builder{
			Descriptors: fds,
		}.Build(),
	))
	return err
}

// MustResetRegistry resets the FauxRPC registry to an empty state. A panic happens if anything fails.
func (c *FauxRPCContainer) MustResetRegistry(ctx context.Context) {
	if err := c.ResetRegistry(ctx); err != nil {
		panic(err)
	}
}

// ResetRegistry resets the FauxRPC registry to an empty state.
func (c *FauxRPCContainer) ResetRegistry(ctx context.Context) error {
	client, err := c.RegistryClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.Reset(ctx, connect.NewRequest(&registryv1.ResetRequest{}))
	return err
}

// MustStubsClient returns a client for managing stubs in the FauxRPC registry. A panic happens if anything fails.
func (c *FauxRPCContainer) MustStubsClient(ctx context.Context) stubsv1connect.StubsServiceClient {
	return stubsv1connect.NewStubsServiceClient(http.DefaultClient, c.MustBaseURL(ctx))
}

// MustStubsClient returns a client for managing stubs in the FauxRPC registry.
func (c *FauxRPCContainer) StubsClient(ctx context.Context) (stubsv1connect.StubsServiceClient, error) {
	baseURL, err := c.BaseURL(ctx)
	if err != nil {
		return nil, err
	}
	return stubsv1connect.NewStubsServiceClient(http.DefaultClient, baseURL), nil
}

// MustAddStub adds a stub to the FauxRPC stub database. A panic happens if anything fails.
func (c *FauxRPCContainer) MustAddStub(ctx context.Context, target string, msg proto.Message) {
	if err := c.AddStub(ctx, target, msg); err != nil {
		panic(err)
	}
}

// AddStub adds a stub to the FauxRPC stub database. Target is the full protobuf path to the type or service method.
//
// Examples:
//
//	err := AddStub(ctx, "connectrpc.eliza.v1.SayResponse", &elizav1.SayResponse{Sentence: "example"})
//	err := AddStub(ctx, "connectrpc.eliza.v1.ElizaService/Say", &elizav1.SayResponse{Sentence: "example"})
func (c *FauxRPCContainer) AddStub(ctx context.Context, target string, msg proto.Message) error {
	client, err := c.StubsClient(ctx)
	if err != nil {
		return err
	}
	msgpb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = client.AddStubs(ctx, connect.NewRequest(
		stubsv1.AddStubsRequest_builder{
			Stubs: []*stubsv1.Stub{
				stubsv1.Stub_builder{
					Ref: stubsv1.StubRef_builder{
						Id:     proto.String(uuid.New().String()),
						Target: proto.String(target),
					}.Build(),
					Proto: msgpb,
				}.Build(),
			},
		}.Build(),
	))
	return err
}

// MustAddStubError adds a stub error response to the FauxRPC stub database. A panic happens if anything fails.
func (c *FauxRPCContainer) MustAddStubError(ctx context.Context, target string, msg string, code stubsv1.ErrorCode) {
	if err := c.AddStubError(ctx, target, msg, code); err != nil {
		panic(err)
	}
}

// AddStubError adds a stub error response to the FauxRPC stub database. A panic happens if anything fails.
//
// Examples:
//
//	err := AddStub(ctx, "connectrpc.eliza.v1.SayResponse", "invalid argument", stubsv1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)
//	err := AddStub(ctx, "connectrpc.eliza.v1.ElizaService/Say", "not found", stubsv1.ErrorCode_ERROR_CODE_NOT_FOUND)
func (c *FauxRPCContainer) AddStubError(ctx context.Context, target string, msg string, code stubsv1.ErrorCode) error {
	client, err := c.StubsClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.AddStubs(ctx, connect.NewRequest(stubsv1.AddStubsRequest_builder{
		Stubs: []*stubsv1.Stub{
			stubsv1.Stub_builder{
				Ref: stubsv1.StubRef_builder{
					Id:     proto.String(uuid.New().String()),
					Target: proto.String(target),
				}.Build(),
				Error: stubsv1.Error_builder{
					Code:    &code,
					Message: proto.String(msg),
				}.Build(),
			}.Build(),
		},
	}.Build()))
	return err
}

// MustResetStubs resets the FauxRPC stub database to an empty state. A panic happens if anything fails.
func (c *FauxRPCContainer) MustResetStubs(ctx context.Context) {
	if err := c.ResetStubs(ctx); err != nil {
		panic(err)
	}
}

// MustResetStubs resets the FauxRPC stub database to an empty state.
func (c *FauxRPCContainer) ResetStubs(ctx context.Context) error {
	client, err := c.StubsClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.RemoveAllStubs(ctx, connect.NewRequest(&stubsv1.RemoveAllStubsRequest{}))
	return err
}
