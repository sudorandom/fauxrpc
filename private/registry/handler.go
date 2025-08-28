package registry

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	registryv1 "github.com/sudorandom/fauxrpc/private/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/private/gen/registry/v1/registryv1connect"
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
	slog.Debug("registry.v1.RegistryService.AddDescriptors()", "files", len(req.Msg.GetDescriptors().GetFile()))
	defer slog.Debug("registry.v1.RegistryService.AddDescriptors() complete")

	files, err := protodesc.NewFiles(req.Msg.GetDescriptors())
	if err != nil {
		return nil, err
	}

	sortedFiles, err := sortFilesByDependency(files)
	if err != nil {
		return nil, err
	}
	for _, fd := range sortedFiles {
		if err := h.registry.RegisterFile(fd); err != nil {
			return nil, fmt.Errorf("%s: %w", fd.FullName(), err)
		}
	}
	if err := h.registry.Rebuild(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&registryv1.AddDescriptorsResponse{}), nil
}

// Reset implements registryv1connect.RegistryServiceHandler.
func (h *handler) Reset(context.Context, *connect.Request[registryv1.ResetRequest]) (*connect.Response[registryv1.ResetResponse], error) {
	slog.Debug("registry.v1.RegistryService.Reset()")
	defer slog.Debug("registry.v1.RegistryService.Reset() complete")

	if err := h.registry.Reset(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&registryv1.ResetResponse{}), nil
}
