package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"
	"github.com/MadAppGang/httplog"
	"github.com/bufbuild/protovalidate-go"
	"github.com/sudorandom/fauxrpc"
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

type ServerOpts struct {
	Version       string
	RenderDocPage bool
	UseReflection bool
	WithHTTPLog   bool
	WithValidate  bool
	OnlyStubs     bool
}

type server struct {
	registry.ServiceRegistry
	stubs.StubDatabase
	lock *sync.Mutex

	handlerOpenAPI          *wrappedHandler
	handlerReflectorV1      *wrappedHandler
	handlerReflectorV1Alpha *wrappedHandler
	handlerTranscoder       *wrappedHandler

	opts ServerOpts
}

func NewServer(opts ServerOpts) (*server, error) {
	serviceRegistry, err := registry.NewServiceRegistry()
	if err != nil {
		return nil, err
	}
	return &server{
		lock:                    &sync.Mutex{},
		ServiceRegistry:         serviceRegistry,
		StubDatabase:            stubs.NewStubDatabase(),
		handlerOpenAPI:          NewWrappedHandler(),
		handlerReflectorV1:      NewWrappedHandler(),
		handlerReflectorV1Alpha: NewWrappedHandler(),
		handlerTranscoder:       NewWrappedHandler(),
		opts:                    opts,
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
	var validate *protovalidate.Validator
	if s.opts.WithValidate {
		v, err := protovalidate.New()
		if err != nil {
			return err
		}
		validate = v
	}

	fakers := []fauxrpc.ProtoFaker{
		stubs.NewStubFaker(s.StubDatabase),
	}
	if !s.opts.OnlyStubs {
		fakers = append(fakers, fauxrpc.NewFauxFaker())
	}

	faker := fauxrpc.NewMultiFaker(fakers)

	s.ServiceRegistry.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
		vgservice := vanguard.NewServiceWithSchema(
			sd, NewHandler(sd, faker, validate),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
		return true
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	s.handlerTranscoder.SetHandler(transcoder)
	if s.opts.UseReflection {
		reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(s.ServiceRegistry.Files()))
		_, v1Handler := grpcreflect.NewHandlerV1(reflector)
		s.handlerReflectorV1.SetHandler(v1Handler)

		_, v1alphaHandler := grpcreflect.NewHandlerV1Alpha(reflector)
		s.handlerReflectorV1Alpha.SetHandler(v1alphaHandler)
	}

	if s.opts.RenderDocPage {
		openapiSpec, err := convertToOpenAPISpec(s.ServiceRegistry, s.opts.Version)
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
	mux.Handle("/", httplog.Logger(s.handlerTranscoder))

	if s.opts.UseReflection {
		mux.Handle("/grpc.reflection.v1.ServerReflection/", httplog.Logger(s.handlerReflectorV1))
		mux.Handle("/grpc.reflection.v1alpha.ServerReflection/", httplog.Logger(s.handlerReflectorV1Alpha))
	}

	// OpenAPI Stuff
	if s.opts.RenderDocPage {
		mux.Handle("GET /fauxrpc/openapi.html", httplog.Logger(singleFileHandler(openapiHTML)))
		mux.Handle("GET /fauxrpc/openapi.yaml", httplog.Logger(s.handlerOpenAPI))
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
