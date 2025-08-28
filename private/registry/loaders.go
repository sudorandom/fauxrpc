package registry

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type LoaderTarget interface {
	protodesc.Resolver
	RegisterFile(file protoreflect.FileDescriptor) error
}

func addServicesFromFileDescriptorProto(registry LoaderTarget, fdp *descriptorpb.FileDescriptorProto) error {
	fd, err := protodesc.NewFile(fdp, registry)
	if err != nil {
		return fmt.Errorf("protodesc.NewFile: %w", err)
	}

	return registry.RegisterFile(fd)
}
