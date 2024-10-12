package registry

import (
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

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

var _ ServiceRegistry = (*serviceRegistry)(nil)

type ServiceRegistry interface {
	Get(name string) protoreflect.ServiceDescriptor
	AddFile(fd protoreflect.FileDescriptor) error
	Reset() error
	ForEachService(cb func(protoreflect.ServiceDescriptor))
	ServiceCount() int
	Files() *protoregistry.Files
}

type serviceRegistry struct {
	services map[string]protoreflect.ServiceDescriptor
	files    *protoregistry.Files
}

func NewServiceRegistry() *serviceRegistry {
	return &serviceRegistry{
		services: map[string]protoreflect.ServiceDescriptor{},
		files:    newFiles(),
	}
}

func (r *serviceRegistry) Reset() error {
	r.services = map[string]protoreflect.ServiceDescriptor{}
	r.files = newFiles()
	return nil
}

func (r *serviceRegistry) Get(name string) protoreflect.ServiceDescriptor {
	return r.services[name]
}

func (r *serviceRegistry) AddFile(fd protoreflect.FileDescriptor) error {
	slog.Debug("add file", "name", fd.FullName(), "path", fd.Path())
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

func (r *serviceRegistry) addService(sd protoreflect.ServiceDescriptor) {
	r.services[string(sd.FullName())] = sd
}

func (r *serviceRegistry) ServiceCount() int {
	return len(r.services)
}

func (r *serviceRegistry) ForEachService(cb func(protoreflect.ServiceDescriptor)) {
	for _, service := range r.services {
		cb(service)
	}
}

func (r *serviceRegistry) Files() *protoregistry.Files {
	return r.files
}

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}
