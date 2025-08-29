package registry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"buf.build/gen/go/grpc/grpc/connectrpc/go/grpc/reflection/v1/reflectionv1connect"
	reflectionv1 "buf.build/gen/go/grpc/grpc/protocolbuffers/go/grpc/reflection/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromReflection uses the given address to connect to a gRPC server that has Server Reflection. The
// services are imported from the file descriptors advertised there.
func AddServicesFromReflection(registry LoaderTarget, httpClient *http.Client, addr string) error {
	slog.Debug("AddServicesFromReflection", slog.String("addr", addr))
	reflectClient := reflectionv1connect.NewServerReflectionClient(httpClient, addr, connect.WithGRPC())
	reflectReq := reflectClient.ServerReflectionInfo(context.Background())
	if err := reflectReq.Send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	}); err != nil {
		return err
	}
	resp, err := reflectReq.Receive()
	if err != nil {
		return err
	}

	var allFdps []*descriptorpb.FileDescriptorProto
	for _, svc := range resp.GetListServicesResponse().GetService() {
		if err := reflectReq.Send(&reflectionv1.ServerReflectionRequest{
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: svc.Name,
			},
		}); err != nil {
			return err
		}
		resp, err := reflectReq.Receive()
		if err != nil {
			return err
		}
		for _, descBytes := range resp.GetFileDescriptorResponse().GetFileDescriptorProto() {
			fdp := new(descriptorpb.FileDescriptorProto)
			if err := proto.Unmarshal(descBytes, fdp); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			allFdps = append(allFdps, fdp)
		}
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

	sortedFiles, err := sortFilesByDependency(files)
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
