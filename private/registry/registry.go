package registry

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var _ ServiceRegistry = (*serviceRegistry)(nil)

type ServiceRegistry interface {
	Get(name string) protoreflect.ServiceDescriptor
	AddFile(fd protoreflect.FileDescriptor) error
	Reset() error
	ForEachService(cb func(protoreflect.ServiceDescriptor))
	ServiceCount() int
	Files() *protoregistry.Files
	NumFiles() int
}

type serviceRegistry struct {
	services     map[string]protoreflect.ServiceDescriptor
	filesOrdered []protoreflect.FileDescriptor
	files        *protoregistry.Files
	lock         *sync.RWMutex
}

func NewServiceRegistry() (*serviceRegistry, error) {
	r := &serviceRegistry{
		services:     map[string]protoreflect.ServiceDescriptor{},
		files:        new(protoregistry.Files),
		filesOrdered: []protoreflect.FileDescriptor{},
		lock:         &sync.RWMutex{},
	}
	return r, AddServicesFromGlobal(r)
}

func (r *serviceRegistry) Reset() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.services = map[string]protoreflect.ServiceDescriptor{}
	r.files = new(protoregistry.Files)
	r.filesOrdered = []protoreflect.FileDescriptor{}
	return AddServicesFromGlobal(r)
}

func (r *serviceRegistry) Get(name string) protoreflect.ServiceDescriptor {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.services[name]
}

func (r *serviceRegistry) AddFile(fd protoreflect.FileDescriptor) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	slog.Debug("AddFile", "name", fd.FullName(), "path", fd.Path())
	if _, err := r.files.FindFileByPath(fd.Path()); err == nil {
		return nil
	} else if !errors.Is(err, protoregistry.NotFound) {
		return err
	}
	if err := r.files.RegisterFile(fd); err != nil {
		if strings.Contains(err.Error(), "name conflict") {
			return nil
		}
		return err
	}
	r.filesOrdered = append(r.filesOrdered, fd)

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
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.services)
}

func (r *serviceRegistry) ForEachService(cb func(protoreflect.ServiceDescriptor)) {
	r.lock.RLock()
	services := r.services
	r.lock.RUnlock()
	for _, service := range services {
		cb(service)
	}
}
func (r *serviceRegistry) ForEachFile(cb func(protoreflect.FileDescriptor)) {
	r.lock.RLock()
	filesOrdered := r.filesOrdered
	r.lock.RUnlock()
	for _, fd := range filesOrdered {
		cb(fd)
	}
}

func (r *serviceRegistry) Files() *protoregistry.Files {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.files
}

func (r *serviceRegistry) NumFiles() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.filesOrdered)
}

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}
