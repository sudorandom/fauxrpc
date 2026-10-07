package registry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/grpcreflect/v2"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromReflection uses the given address to connect to a gRPC server that has Server Reflection. The
// services are imported from the file descriptors advertised there.
func AddServicesFromReflection(registry LoaderTarget, httpClient *http.Client, addr string) error {
	slog.Debug("AddServicesFromReflection", slog.String("addr", addr))
	client := grpcreflect.NewClient(connect.NewClient(
		connecthttp.NewTransport(httpClient, addr, connecthttp.WithGRPC()),
	))
	stream := client.NewStream(context.Background())
	defer stream.Close()

	services, err := stream.ListServices()
	if err != nil {
		return err
	}

	var allFdps []*descriptorpb.FileDescriptorProto
	for _, svc := range services {
		fdps, err := stream.FileContainingSymbol(svc)
		if err != nil {
			return err
		}
		allFdps = append(allFdps, fdps...)
	}

	// Deduplicate file descriptors. It's possible that we get the same file
	// descriptor multiple times if we ask for multiple services that are in the
	// same file or have common dependencies.
	seen := make(map[string]struct{})
	uniqueFdps := make([]*descriptorpb.FileDescriptorProto, 0, len(allFdps))
	for _, fdp := range allFdps {
		if fdp.Name == nil {
			continue
		}
		if _, ok := seen[fdp.GetName()]; !ok {
			seen[fdp.GetName()] = struct{}{}
			uniqueFdps = append(uniqueFdps, fdp)
		}
	}

	fds := &descriptorpb.FileDescriptorSet{
		File: uniqueFdps,
	}
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return fmt.Errorf("protodesc.NewFiles: %w", err)
	}

	sortedFiles, err := SortFilesByDependency(files)
	if err != nil {
		return fmt.Errorf("sortFilesByDependency: %w", err)
	}

	for _, fd := range sortedFiles {
		if err := registry.RegisterFile(fd); err != nil {
			return fmt.Errorf("RegisterFile: %w", err)
		}
	}

	return nil
}
