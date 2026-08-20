package registry

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

var _ ServiceRegistry = (*serviceRegistry)(nil)

type ServiceRegistry interface {
	Get(name string) protoreflect.ServiceDescriptor
	Reset() error
	ForEachService(cb func(protoreflect.ServiceDescriptor) bool)
	ServiceCount() int
	Files() *protoregistry.Files
	Types() *protoregistry.Types
	Resolver() *Resolver
	NumFiles() int
	Rebuild() error
	// Act like protoreflect.Files
	RegisterFile(file protoreflect.FileDescriptor) error
	// protoregistry.Resolver
	FindFileByPath(string) (protoreflect.FileDescriptor, error)
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
}

type serviceRegistry struct {
	services     map[string]protoreflect.ServiceDescriptor
	filesOrdered []protoreflect.FileDescriptor
	files        *protoregistry.Files
	types        *protoregistry.Types
	lock         *sync.RWMutex
}

func NewServiceRegistry() (*serviceRegistry, error) {
	r := &serviceRegistry{
		services:     map[string]protoreflect.ServiceDescriptor{},
		files:        new(protoregistry.Files),
		types:        new(protoregistry.Types),
		filesOrdered: []protoreflect.FileDescriptor{},
		lock:         &sync.RWMutex{},
	}
	return r, AddServicesFromGlobal(r)
}

func (r *serviceRegistry) Reset() error {
	slog.Debug("serviceRegistry.Reset()")
	defer slog.Debug("serviceRegistry.Reset() complete")

	r.lock.Lock()
	r.services = map[string]protoreflect.ServiceDescriptor{}
	r.files = new(protoregistry.Files)
	r.types = new(protoregistry.Types)
	r.filesOrdered = []protoreflect.FileDescriptor{}
	r.lock.Unlock()
	return AddServicesFromGlobal(r)
}

func (r *serviceRegistry) Get(name string) protoreflect.ServiceDescriptor {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.services[name]
}

func (r *serviceRegistry) Rebuild() error {
	return nil
}

func (r *serviceRegistry) RegisterFile(fd protoreflect.FileDescriptor) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	slog.Debug("serviceRegistry.RegisterFile()", "name", fd.FullName(), "path", fd.Path())
	defer slog.Debug("serviceRegistry.RegisterFile() complete", "name", fd.FullName(), "path", fd.Path())

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
	r.addExtensions(fd)

	svcs := fd.Services()
	for i := 0; i < svcs.Len(); i++ {
		svc := svcs.Get(i)
		r.addService(svc)
	}
	return nil
}

// addExtensions records every extension the file declares, at file scope and
// inside its messages. Nothing else knows they exist: an extension is reachable
// only by looking it up by the message it extends, which needs a type registry
// to look in.
func (r *serviceRegistry) addExtensions(container extensionContainer) {
	exts := container.Extensions()
	for i := 0; i < exts.Len(); i++ {
		xd := exts.Get(i)
		xt := extensionType(xd)
		if err := r.types.RegisterExtension(xt); err != nil {
			// A conflict means some other file already claimed this name, which
			// is the same situation RegisterFile tolerates above.
			slog.Debug("skipping extension", "name", xd.FullName(), "error", err)
		}
	}
	msgs := container.Messages()
	for i := 0; i < msgs.Len(); i++ {
		r.addExtensions(msgs.Get(i))
	}
}

// extensionContainer is what files and messages have in common: both can
// declare extensions and hold nested messages that declare more.
type extensionContainer interface {
	Extensions() protoreflect.ExtensionDescriptors
	Messages() protoreflect.MessageDescriptors
}

// extensionType prefers the extension compiled into this binary, so that
// setting it on a generated message uses that message's own Go type. Schemas
// loaded at runtime have no Go type and fall back to a dynamic one.
func extensionType(xd protoreflect.ExtensionDescriptor) protoreflect.ExtensionType {
	if xt, err := protoregistry.GlobalTypes.FindExtensionByName(xd.FullName()); err == nil {
		return xt
	}
	return dynamicpb.NewExtensionType(xd)
}

func (r *serviceRegistry) addService(sd protoreflect.ServiceDescriptor) {
	r.services[string(sd.FullName())] = sd
}

func (r *serviceRegistry) ServiceCount() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.services)
}

func (r *serviceRegistry) ForEachService(cb func(protoreflect.ServiceDescriptor) bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, service := range r.services {
		if !cb(service) {
			break
		}
	}
}
func (r *serviceRegistry) ForEachFile(cb func(protoreflect.FileDescriptor)) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, fd := range r.filesOrdered {
		cb(fd)
	}
}

func (r *serviceRegistry) Files() *protoregistry.Files {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.files
}

// Types holds the extensions declared by every registered file. It doubles as
// the resolver protojson needs to read back JSON that names an extension.
func (r *serviceRegistry) Types() *protoregistry.Types {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.types
}

func (r *serviceRegistry) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.files.FindFileByPath(path)
}

func (r *serviceRegistry) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.files.FindDescriptorByName(name)
}

func (r *serviceRegistry) NumFiles() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.filesOrdered)
}

func NewMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}
