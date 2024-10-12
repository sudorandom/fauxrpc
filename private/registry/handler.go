package registry

import (
	"context"

	"connectrpc.com/connect"
	registryv1 "github.com/sudorandom/fauxrpc/proto/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	"google.golang.org/protobuf/reflect/protodesc"
)

var _ registryv1connect.RegistryServiceHandler = (*handler)(nil)

type handler struct {
	registry ServiceRegistry
}

func NewHandler(registry ServiceRegistry) *handler {
	return &handler{registry: registry}
}

// AddDescriptors implements registryv1connect.RegistryServiceHandler.
func (h *handler) AddDescriptors(ctx context.Context, req *connect.Request[registryv1.AddDescriptorsRequest]) (*connect.Response[registryv1.AddDescriptorsResponse], error) {
	for _, fdp := range req.Msg.Descriptors.File {
		fd, err := protodesc.NewFile(fdp, h.registry.Files())
		if err != nil {
			return nil, err
		}
		if err := h.registry.AddFile(fd); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&registryv1.AddDescriptorsResponse{}), nil
}

// Reset implements registryv1connect.RegistryServiceHandler.
func (h *handler) Reset(context.Context, *connect.Request[registryv1.ResetRequest]) (*connect.Response[registryv1.ResetResponse], error) {
	if err := h.registry.Reset(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&registryv1.ResetResponse{}), nil
}
