package registry

import (
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromDescriptorsFilePB imports services from a .pb file
func AddServicesFromDescriptorsFilePB(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromDescriptorsFilePB", slog.String("filepath", filepath))
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptorpb.FileDescriptorSet)
	if err := proto.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromFileDescriptorProto(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}
