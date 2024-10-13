package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ registry.ServiceRegistry = (*server)(nil)
var _ stubs.StubDatabase = (*server)(nil)

type Server interface {
	registry.ServiceRegistry
	stubs.StubDatabase
}

type server struct {
	registry.ServiceRegistry
	stubs.StubDatabase
	lock *sync.Mutex

	handlerOpenAPI          *wrappedHandler
	handlerReflectorV1      *wrappedHandler
	handlerReflectorV1Alpha *wrappedHandler
	handlerTranscoder       *wrappedHandler

	version       string
	renderDocPage bool
	useReflection bool
}

func NewServer(version string, renderDocPage, useReflection bool) (*server, error) {
	serviceRegistry, err := registry.NewServiceRegistry()
	if err != nil {
		return nil, err
	}
	return &server{
		lock:                    &sync.Mutex{},
		ServiceRegistry:         serviceRegistry,
		StubDatabase:            stubs.NewStubDatabase(),
		version:                 version,
		renderDocPage:           renderDocPage,
		useReflection:           useReflection,
		handlerOpenAPI:          NewWrappedHandler(),
		handlerReflectorV1:      NewWrappedHandler(),
		handlerReflectorV1Alpha: NewWrappedHandler(),
		handlerTranscoder:       NewWrappedHandler(),
	}, nil
}

func (s *server) Reset() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ServiceRegistry.Reset(); err != nil {
		return err
	}
	return s.rebuildHandlers()
}

func (s *server) AddFile(fd protoreflect.FileDescriptor) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ServiceRegistry.AddFile(fd); err != nil {
		return err
	}
	return s.rebuildHandlers()
}

func (s *server) AddFileFrompath(path string) error {
	return registry.AddServicesFromPath(s.ServiceRegistry, path)
}

func (s *server) rebuildHandlers() error {
	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	s.ServiceRegistry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		vgservice := vanguard.NewServiceWithSchema(
			sd, NewHandler(sd, s.StubDatabase),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	s.handlerTranscoder.SetHandler(transcoder)
	if s.useReflection {
		reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(s.ServiceRegistry.Files()))
		_, v1Handler := grpcreflect.NewHandlerV1(reflector)
		s.handlerReflectorV1.SetHandler(v1Handler)

		_, v1alphaHandler := grpcreflect.NewHandlerV1Alpha(reflector)
		s.handlerReflectorV1.SetHandler(v1alphaHandler)
	}

	if s.renderDocPage {
		openapiSpec, err := convertToOpenAPISpec(s.ServiceRegistry, s.version)
		if err != nil {
			return err
		}
		s.handlerOpenAPI.SetHandler(singleFileHandler(string(openapiSpec)))
	}
	return nil
}

func (s *server) Mux() (*http.ServeMux, error) {
	if err := s.rebuildHandlers(); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", s.handlerTranscoder)

	if s.useReflection {
		mux.Handle("/grpc.reflection.v1.ServerReflection/", s.handlerReflectorV1)
		mux.Handle("/grpc.reflection.v1alpha.ServerReflection/", s.handlerReflectorV1Alpha)
	}

	// OpenAPI Stuff
	if s.renderDocPage {
		mux.Handle("GET /fauxrpc/openapi.html", singleFileHandler(openapiHTML))
		mux.Handle("GET /fauxrpc/openapi.yaml", s.handlerOpenAPI)
	}
	return mux, nil
}

type staticNames struct {
	names []string
}

func (n *staticNames) Names() []string {
	return n.names
}

func singleFileHandler(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}
}

type wrappedHandler struct {
	lock    *sync.RWMutex
	handler http.Handler
}

func NewWrappedHandler() *wrappedHandler {
	return &wrappedHandler{
		lock:    &sync.RWMutex{},
		handler: http.NotFoundHandler(),
	}
}

func (h *wrappedHandler) SetHandler(handler http.Handler) {
	h.lock.Lock()
	h.handler = handler
	h.lock.Unlock()
}

func (h *wrappedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	h.handler.ServeHTTP(w, r)
}
