package registry

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"buf.build/gen/go/grpc/grpc/connectrpc/go/grpc/reflection/v1/reflectionv1connect"
	reflectionv1 "buf.build/gen/go/grpc/grpc/protocolbuffers/go/grpc/reflection/v1"
	"buf.build/go/protoyaml"
	"connectrpc.com/connect"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	// Ensure the wellknown types get imported and registered into the global registry
	anypb "google.golang.org/protobuf/types/known/anypb"
	apipb "google.golang.org/protobuf/types/known/apipb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
	sourcecontextpb "google.golang.org/protobuf/types/known/sourcecontextpb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	typepb "google.golang.org/protobuf/types/known/typepb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type LoaderTarget interface {
	protodesc.Resolver
	RegisterFile(file protoreflect.FileDescriptor) error
}

// AddServicesFromPath imports services from a given 'path' which can be a local file path, directory,
// BSR repo, server address for server reflection.
func AddServicesFromPath(registry LoaderTarget, path string) error {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return AddServicesFromReflection(registry, http.DefaultClient, path)
	}
	stat, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) && looksLikeBSR(path) {
		return AddServicesFromBSR(registry, path)
	} else if err != nil {
		return err
	}
	if stat.IsDir() {
		if err := fs.WalkDir(os.DirFS(path), ".", func(childpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if err := AddServicesFromSingleFile(registry, filepath.Join(path, childpath)); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return AddServicesFromSingleFile(registry, path)
}

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

// AddServicesFromDescriptorsFilePB imports services from a .proto file
func AddServicesFromProtoFile(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromProtoFile", slog.String("filepath", filepath))
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	handler := reporter.NewHandler(nil)
	ast, err := parser.Parse(filepath, f, handler)
	if err != nil {
		return err
	}
	res, err := parser.ResultFromAST(ast, true, handler)
	if err != nil {
		return fmt.Errorf("convert from AST: %w", err)
	}
	fd, err := protodesc.NewFile(res.FileDescriptorProto(), registry)
	if err != nil {
		return fmt.Errorf("protodesc.NewFile: %w", err)
	}

	return registry.RegisterFile(fd)
}

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

// AddServicesFromDescriptorsFileTXTPB imports services from a .txtpb file
func AddServicesFromDescriptorsFileTXTPB(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromDescriptorsFileTXTPB", slog.String("filepath", filepath))
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptorpb.FileDescriptorSet)
	unmarshaller := prototext.UnmarshalOptions{DiscardUnknown: true}
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

	fds := &descriptorpb.FileDescriptorSet{
		File: allFdps,
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

func addServicesFromFileDescriptorProto(registry LoaderTarget, fdp *descriptorpb.FileDescriptorProto) error {
	fd, err := protodesc.NewFile(fdp, registry)
	if err != nil {
		return fmt.Errorf("protodesc.NewFile: %w", err)
	}

	return registry.RegisterFile(fd)
}

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}

// AddServicesFromBSR adds services from the BSR. Not yet supported
func AddServicesFromBSR(registry LoaderTarget, module string) error {
	slog.Debug("AddServicesFromBSR", slog.String("module", module))
	// TODO: Add support for downloading from the buf registry.
	// buf.registry.module.v1.CommitService/GetCommits
	// buf.registry.module.v1.GraphService/GetGraph
	// buf.registry.module.v1.ModuleService/GetModules
	// buf.registry.owner.v1.OwnerService/GetOwners
	return errors.New("BSR is not (yet) supported")
}

// AddServicesFromGlobal adds the 'well known' types to the registry. This is typically implicitly called.
func AddServicesFromGlobal(registry LoaderTarget) error {
	for _, fd := range []protoreflect.FileDescriptor{
		descriptorpb.File_google_protobuf_descriptor_proto,
		anypb.File_google_protobuf_any_proto,
		apipb.File_google_protobuf_api_proto,
		durationpb.File_google_protobuf_duration_proto,
		emptypb.File_google_protobuf_empty_proto,
		fieldmaskpb.File_google_protobuf_field_mask_proto,
		sourcecontextpb.File_google_protobuf_source_context_proto,
		structpb.File_google_protobuf_struct_proto,
		timestamppb.File_google_protobuf_timestamp_proto,
		typepb.File_google_protobuf_type_proto,
		wrapperspb.File_google_protobuf_wrappers_proto,
	} {
		if err := registry.RegisterFile(fd); err != nil {
			return err
		}
	}
	return nil
}
