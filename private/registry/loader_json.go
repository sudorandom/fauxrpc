package registry

import (
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromDescriptorsFileJSON imports services from a .json file
func AddServicesFromDescriptorsFileJSON(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromDescriptorsFileJSON", slog.String("filepath", filepath))
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptorpb.FileDescriptorSet)
	unmarshaller := protojson.UnmarshalOptions{DiscardUnknown: true}
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
