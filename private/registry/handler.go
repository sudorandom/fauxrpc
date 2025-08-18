package registry

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	registryv1 "github.com/sudorandom/fauxrpc/proto/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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

func sortFilesByDependency(files *protoregistry.Files) ([]protoreflect.FileDescriptor, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Build the dependency graph.
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		inDegree[fd.Path()] = 0
		imports := fd.Imports()
		for i := 0; i < imports.Len(); i++ {
			imp := imports.Get(i)
			graph[imp.Path()] = append(graph[imp.Path()], fd.Path())
			inDegree[fd.Path()]++
		}
		return true
	})

	// Topological sort using Kahn's algorithm.
	var queue []string
	for fileName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, fileName)
		}
	}

	var sortedFiles []protoreflect.FileDescriptor
	for len(queue) > 0 {
		currentFile := queue[0]
		queue = queue[1:]

		fd, err := files.FindFileByPath(currentFile)
		if err != nil {
			return nil, fmt.Errorf("failed to find file %q: %v", currentFile, err)
		}
		sortedFiles = append(sortedFiles, fd)

		for _, neighbor := range graph[currentFile] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return sortedFiles, nil
}
