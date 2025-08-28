package registry

import (
	"log/slog"
	"path"
)

// AddServicesFromSingleFile imports services from a single (non-directory) file
func AddServicesFromSingleFile(registry LoaderTarget, filepath string) error {
	ext := path.Ext(filepath)
	switch ext {
	case ".proto":
		return AddServicesFromProtoFile(registry, filepath)
	case ".txtpb":
		return AddServicesFromDescriptorsFileTXTPB(registry, filepath)
	case ".json":
		return AddServicesFromDescriptorsFileJSON(registry, filepath)
	case ".yaml":
		return AddServicesFromDescriptorsFileYAML(registry, filepath)
	case ".binpb":
		return AddServicesFromDescriptorsFilePB(registry, filepath)
	default:
		slog.Info("not sure how to handle file", "filepath", filepath)
	}
	return nil
}
