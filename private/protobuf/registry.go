package protobuf

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
	"connectrpc.com/connect"
	"github.com/bufbuild/protoyaml-go"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	// Ensure the wellknown types get imported and registered into the global registry
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "google.golang.org/protobuf/types/known/apipb"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	_ "google.golang.org/protobuf/types/known/sourcecontextpb"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "google.golang.org/protobuf/types/known/typepb"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
)

type ServiceRegistry struct {
	services map[string]protoreflect.ServiceDescriptor
	files    *protoregistry.Files
}

func NewServiceRegistry() *ServiceRegistry {
	files := new(protoregistry.Files)
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		files.RegisterFile(fd)
		return true
	})

	return &ServiceRegistry{
		services: map[string]protoreflect.ServiceDescriptor{},
		files:    files,
	}
}

func (r *ServiceRegistry) Get(name string) protoreflect.ServiceDescriptor {
	return r.services[name]
}

func (r *ServiceRegistry) AddFile(fd protoreflect.FileDescriptor) error {
	slog.Info("add file", "name", fd.FullName(), "path", fd.Path())
	if _, err := r.files.FindFileByPath(fd.Path()); err == nil {
		return nil
	} else if !errors.Is(err, protoregistry.NotFound) {
		return err
	}
	if err := r.files.RegisterFile(fd); err != nil {
		return err
	}

	svcs := fd.Services()
	for i := 0; i < svcs.Len(); i++ {
		svc := svcs.Get(i)
		r.addService(svc)
	}
	return nil
}

func (r *ServiceRegistry) addService(sd protoreflect.ServiceDescriptor) {
	r.services[string(sd.FullName())] = sd
}

func (r *ServiceRegistry) ServiceCount() int {
	return len(r.services)
}

func (r *ServiceRegistry) ForEachService(cb func(protoreflect.ServiceDescriptor)) {
	for _, service := range r.services {
		cb(service)
	}
}

func (r *ServiceRegistry) Files() protodesc.Resolver {
	return r.files
}

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}

func AddServicesFromPath(registry *ServiceRegistry, path string) error {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return AddServicesFromReflection(registry, path)
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

func AddServicesFromSingleFile(registry *ServiceRegistry, filepath string) error {
	slog.Info("AddServicesFromSingleFile", "filepath", filepath)
	ext := path.Ext(filepath)
	switch ext {
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

func AddServicesFromDescriptorsFilePB(registry *ServiceRegistry, filepath string) error {
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptor.FileDescriptorSet)
	if err := proto.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromDescriptorsBytes(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}

func AddServicesFromDescriptorsFileJSON(registry *ServiceRegistry, filepath string) error {
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptor.FileDescriptorSet)
	unmarshaller := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaller.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromDescriptorsBytes(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}

func AddServicesFromDescriptorsFileYAML(registry *ServiceRegistry, filepath string) error {
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptor.FileDescriptorSet)
	unmarshaller := protoyaml.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaller.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromDescriptorsBytes(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}

func AddServicesFromDescriptorsFileTXTPB(registry *ServiceRegistry, filepath string) error {
	descBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	dss := new(descriptor.FileDescriptorSet)
	unmarshaller := prototext.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaller.Unmarshal(descBytes, dss); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, fdp := range dss.File {
		if err := addServicesFromDescriptorsBytes(registry, fdp); err != nil {
			return err
		}
	}
	return nil
}

func AddServicesFromReflection(registry *ServiceRegistry, addr string) error {
	reflectClient := reflectionv1connect.NewServerReflectionClient(http.DefaultClient, addr, connect.WithGRPC())
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
			fdp := new(descriptor.FileDescriptorProto)
			if err := proto.Unmarshal(descBytes, fdp); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}

			if err := addServicesFromDescriptorsBytes(registry, fdp); err != nil {
				return err
			}
		}
	}
	return nil
}

func addServicesFromDescriptorsBytes(registry *ServiceRegistry, fdp *descriptorpb.FileDescriptorProto) error {
	slog.Debug("addServicesFromDescriptorsBytes", slog.String("fdp", fdp.GetName()))
	fd, err := protodesc.NewFile(fdp, registry.Files())
	if err != nil {
		return fmt.Errorf("protodesc.NewFile: %w", err)
	}

	return registry.AddFile(fd)
}

func AddServicesFromBSR(registry *ServiceRegistry, module string) error {
	// TODO: Add support for downloading from the buf registry.
	// It might just be easier to use a library for this.
	// https://buf.build/bufbuild/registry/docs/main:buf.registry.module.v1#buf.registry.module.v1.File
	// call buf.registry.module.v1.ModuleService/GetModules
	// parts := strings.Split(module, "/")
	// client := modulev1connect.NewModuleServiceClient(http.DefaultClient, "https://buf.build/")
	// modules, err := client.GetModules(context.Background(), connect.NewRequest(&modulev1.GetModulesRequest{
	// 	ModuleRefs: []*modulev1.ModuleRef{
	// 		{
	// 			Value: &modulev1.ModuleRef_Name_{
	// 				Name: &modulev1.ModuleRef_Name{
	// 					Owner:  parts[1],
	// 					Module: parts[2],
	// 				},
	// 			},
	// 		},
	// 	},
	// }))
	// if err != nil {
	// 	return err
	// }
	// log.Println(modules.Msg)
	// dlclient := modulev1connect.NewDownloadServiceClient(http.DefaultClient, "https://buf.build/")
	// dlclient.Download(context.Background(), connect.NewRequest(&modulev1.DownloadRequest{
	// 	Values: []*modulev1.DownloadRequest_Value{
	// 		{
	// 			ResourceRef:        &modulev1.ResourceRef{},
	// 			FileTypes:          []modulev1.FileType{},
	// 			Paths:              []string{},
	// 			PathsAllowNotExist: false,
	// 		},
	// 	},
	// }))
	// buf.registry.module.v1.CommitService/GetCommits
	// buf.registry.module.v1.GraphService/GetGraph
	// buf.registry.module.v1.ModuleService/GetModules
	// buf.registry.owner.v1.OwnerService/GetOwners
	return errors.New("BSR is not (yet) supported")
}
