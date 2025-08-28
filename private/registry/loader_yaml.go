package registry

import (
	"fmt"
	"log/slog"
	"os"

	"buf.build/go/protoyaml"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromDescriptorsFileYAML imports services from a .yaml file
func AddServicesFromDescriptorsFileYAML(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromDescriptorsFileYAML", slog.String("filepath", filepath))
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptorpb.FileDescriptorSet)
	unmarshaller := protoyaml.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaller.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromFileDescriptorProto(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}
