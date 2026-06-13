package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"github.com/MadAppGang/httplog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/frontend"
	"github.com/sudorandom/fauxrpc/private/gen/registry/v1/registryv1connect"
	"github.com/sudorandom/fauxrpc/private/gen/stubs/v1/stubsv1connect"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/protodocs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

var _ registry.ServiceRegistry = (*server)(nil)
var _ stubs.StubDatabase = (*server)(nil)

type Server interface {
	registry.ServiceRegistry
	stubs.StubDatabase
	GetStats() *metrics.Stats
	IncrementTotalRequests()
	IncrementErrors()
	GetLogger() *fauxlog.Logger
	GetMaxDepth() int
	GetProxyTo() string
	GetRecordDir() string
	GetProxyClient() *http.Client
}

type ServerOpts struct {
	Version       string
	RenderDocPage bool
	UseReflection bool
	WithHTTPLog   bool
	WithValidate  bool
	OnlyStubs     bool
	NoCORS        bool
	Addr          string
	WithDashboard bool
	MaxDepth      int
	ProxyTo       string
	RecordDir     string
}

type server struct {
	registry.ServiceRegistry
	stubs.StubDatabase
	lock *sync.Mutex

	handlerDocs             *wrappedHandler
	handlerReflectorV1      *wrappedHandler
	handlerReflectorV1Alpha *wrappedHandler
	handlerTranscoder       *wrappedHandler

	opts        ServerOpts
	stats       *metrics.Stats
	logger      *fauxlog.Logger
	proxyClient *http.Client
}

func NewServer(opts ServerOpts) (*server, error) {
	serviceRegistry, err := registry.NewServiceRegistry()
	if err != nil {
		return nil, err
	}
	s := &server{
		lock:                    &sync.Mutex{},
		ServiceRegistry:         serviceRegistry,
		StubDatabase:            stubs.NewStubDatabase(),
		handlerDocs:             NewWrappedHandler(),
		handlerReflectorV1:      NewWrappedHandler(),
		handlerReflectorV1Alpha: NewWrappedHandler(),
		handlerTranscoder:       NewWrappedHandler(),
		opts:                    opts,
		stats: &metrics.Stats{
			StartedAt:      time.Now(),
			LastReset:      time.Now(),
			HTTPHost:       opts.Addr,
			GoVersion:      runtime.Version(),
			FauxRpcVersion: opts.Version,
			RequestCounts:  make(map[time.Time]int64),
		},
		logger:      fauxlog.NewLogger(),
		proxyClient: newProxyClient(),
	}
	if opts.RecordDir != "" {
		go s.startRecordDirWorker()
	}
	return s, nil
}

func (s *server) Reset() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ServiceRegistry.Reset(); err != nil {
		return err
	}
	s.stats.Reset()
	return s.rebuildHandlers()
}

func (s *server) Rebuild() error {
	return s.rebuildHandlers()
}

func (s *server) RegisterFile(fd protoreflect.FileDescriptor) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ServiceRegistry.RegisterFile(fd); err != nil {
		return err
	}
	return nil
}

func (s *server) AddFileFromPath(ctx context.Context, path string) error {
	return registry.AddServicesFromPath(ctx, s.ServiceRegistry, path)
}

func (s *server) GetLogger() *fauxlog.Logger {
	return s.logger
}

func (s *server) GetStubDB() stubs.StubDatabase {
	return s.StubDatabase
}

func (s *server) GetStats() *metrics.Stats {
	stats := s.stats.Copy()

	stats.UniqueServices = s.ServiceCount()
	uniqueMethods := make(map[string]struct{})
	s.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
		methods := sd.Methods()
		for i := 0; i < methods.Len(); i++ {
			method := methods.Get(i)
			uniqueMethods[string(method.FullName())] = struct{}{}
		}
		return true
	})
	stats.UniqueMethods = int(len(uniqueMethods))
	stats.HTTPHost = s.opts.Addr
	stats.FauxRpcVersion = s.opts.Version
	// Calculate requests per second for the last second
	now := time.Now().Truncate(time.Second)
	lastSecond := now.Add(-time.Second)
	stats.RequestsPerSecond = stats.RequestCounts[lastSecond]

	// Clean up old entries
	for t := range stats.RequestCounts {
		if t.Before(lastSecond) { // Clean up anything older than the last full second
			delete(stats.RequestCounts, t)
		}
	}

	if stats.TotalRequests > 0 {
		stats.ErrorRate = fmt.Sprintf("%.3f%%", float64(stats.Errors)/float64(stats.TotalRequests)*100)
	} else {
		stats.ErrorRate = "0.000%"
	}

	return stats
}

func (s *server) IncrementTotalRequests() {
	s.stats.IncrementTotalRequests()
}

func (s *server) IncrementErrors() {
	s.stats.IncrementErrors()
}

func (s *server) GetMaxDepth() int {
	return s.opts.MaxDepth
}

func (s *server) GetProxyTo() string {
	return s.opts.ProxyTo
}

func (s *server) GetRecordDir() string {
	return s.opts.RecordDir
}

func (s *server) GetProxyClient() *http.Client {
	return s.proxyClient
}

func (s *server) rebuildHandlers() error {
	slog.Debug("Rebuilding handlers")
	defer slog.Debug("Rebuilding handlers complete")

	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	var validate protovalidate.Validator
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

	s.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
		maxDepth := s.opts.MaxDepth
		if maxDepth <= 0 {
			maxDepth = 5
		}
		vgservice := vanguard.NewServiceWithSchema(
			sd, NewHandler(sd, faker, validate, s, s.logger, maxDepth), // Pass the server instance here
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
		reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(s.ServiceRegistry))
		_, v1Handler := grpcreflect.NewHandlerV1(reflector)
		s.handlerReflectorV1.SetHandler(v1Handler)

		_, v1alphaHandler := grpcreflect.NewHandlerV1Alpha(reflector)
		s.handlerReflectorV1Alpha.SetHandler(v1alphaHandler)
	}

	if s.opts.RenderDocPage {
		descBuilder := strings.Builder{}
		descBuilder.WriteString("# FauxRPC Service Documentation\n\n")
		descBuilder.WriteString("This is a [FauxRPC](https://fauxrpc.com/) server that is currently hosting the following services:\n\n")
		s.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
			descBuilder.WriteString("- ")
			descBuilder.WriteString(string(sd.FullName()))
			descBuilder.WriteByte('\n')
			return true
		})
		descBuilder.WriteString("\nFauxRPC is a mock server that supports gRPC, gRPC-Web, Connect, and HTTP/JSON transcoding.")

		sortedFiles, err := registry.SortFilesByDependency(s.Files())
		if err != nil {
			return err
		}
		fds := &descriptorpb.FileDescriptorSet{}
		for _, fd := range sortedFiles {
			fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fd))
		}

		handler, err := protodocs.NewHandler(protodocs.Config{
			Title:                     "FauxRPC",
			LogoText:                  "FauxRPC",
			Descriptors:               fds,
			Prefix:                    "/fauxrpc/docs/",
			FrontPageMarkdown:         descBuilder.String(),
			BottomOfFrontPageMarkdown: "FauxRPC is open source and available at [github.com/sudorandom/fauxrpc](https://github.com/sudorandom/fauxrpc).",
			BackToText:                "Back to Dashboard",
			BackToURL:                 "/fauxrpc",
		})
		if err != nil {
			return err
		}
		s.handlerDocs.SetHandler(handler)
	}
	return nil
}

func (s *server) Handler() (http.Handler, error) {
	if err := s.rebuildHandlers(); err != nil {
		return nil, err
	}

	mux := chi.NewMux()
	mux.Use(middleware.RequestID, middleware.Recoverer, httplog.Logger)
	if !s.opts.NoCORS {
		mux.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: connectcors.AllowedMethods(),
			AllowedHeaders: connectcors.AllowedHeaders(),
			ExposedHeaders: connectcors.ExposedHeaders(),
		}).Handler)
	}

	mux.Mount("/", protocolMiddleware(s.handlerTranscoder))
	if s.opts.WithDashboard {
		mux.Handle("/", http.RedirectHandler("/fauxrpc", http.StatusFound))
		mux.Handle("/fauxrpc/assets/", http.StripPrefix("/fauxrpc/assets/", http.FileServer(http.Dir("private/frontend/assets"))))
		mux.Mount("/fauxrpc", frontend.DashboardHandler(s))
	}

	if s.opts.UseReflection {
		mux.Mount("/grpc.reflection.v1.ServerReflection/", s.handlerReflectorV1)
		mux.Mount("/grpc.reflection.v1alpha.ServerReflection/", s.handlerReflectorV1Alpha)
	}

	// Docs Stuff
	if s.opts.RenderDocPage {
		mux.Get("/fauxrpc/docs", http.RedirectHandler("/fauxrpc/docs/", http.StatusMovedPermanently).ServeHTTP)
		mux.Handle("/fauxrpc/docs/*", s.handlerDocs)
	}

	validateInterceptor := validate.NewInterceptor()
	mux.Mount(stubsv1connect.NewStubsServiceHandler(stubs.NewHandler(s, s), connect.WithInterceptors(validateInterceptor)))
	mux.Mount(registryv1connect.NewRegistryServiceHandler(registry.NewHandler(s), connect.WithInterceptors(validateInterceptor)))

	return mux, nil
}

type staticNames struct {
	names []string
}

func (n *staticNames) Names() []string {
	return n.names
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

func protocolMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		protocol := getClientProtocol(r)
		ctx := context.WithValue(r.Context(), clientProtocolKey, protocol)

		headers, _ := json.Marshal(maskHeaders(r.Header))
		ctx = context.WithValue(ctx, requestHeadersKey, headers)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getClientProtocol(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	connectVersion := r.Header.Get("Connect-Protocol-Version")
	// Connect-Protocol-Version
	switch {
	case strings.HasPrefix(contentType, "application/grpc-web"):
		return "gRPC-Web"
	case strings.HasPrefix(contentType, "application/grpc"):
		return "gRPC"
	case strings.HasPrefix(contentType, "application/connect"):
		return "ConnectRPC"
	case connectVersion != "":
		return "ConnectRPC"
	}
	return "HTTP"
}

func maskHeaders(headers http.Header) http.Header {
	maskedHeaders := headers.Clone()
	for _, h := range []string{
		"Authorization",
		"Proxy-Authorization",
		"Proxy-Authenticate",
		"WWW-Authenticate",
		"X-API-Key",
		"X-Auth-Token",
		"Cookie",
	} {
		if maskedHeaders.Get(h) != "" {
			maskedHeaders.Set(h, "*****")
		}
	}
	return maskedHeaders
}

func (s *server) startRecordDirWorker() {
	ch, unsubscribe := s.logger.Subscribe()
	defer unsubscribe()

	for entry := range ch {
		// 1. Check if the request was proxied
		var headers http.Header
		if err := json.Unmarshal(entry.ResponseHeaders, &headers); err != nil {
			continue
		}
		if headers.Get("X-Fauxrpc-Source") == "" {
			continue
		}

		// 2. Locate the service and method descriptor
		sd := s.Get(entry.Service)
		if sd == nil {
			continue
		}
		method := sd.Methods().ByName(protoreflect.Name(entry.Method))
		if method == nil {
			continue
		}

		// 3. Construct target and filepath
		target := entry.Service + "/" + entry.Method
		filePath := filepath.Join(s.opts.RecordDir, target+".json")

		// 4. Determine if it is a streaming request or response
		isClientStream := method.IsStreamingClient()
		isServerStream := method.IsStreamingServer()

		// 5. Generate active_if from the request(s)
		var activeIf string
		if !isClientStream {
			if len(entry.RequestBody) > 0 {
				reqMsg := dynamicpb.NewMessage(method.Input())
				if err := protojson.Unmarshal(entry.RequestBody, reqMsg); err == nil {
					activeIf = stubs.ActiveIfFromProto(reqMsg)
				}
			}
		} else {
			if len(entry.RequestFrames) > 0 {
				reqMsg := dynamicpb.NewMessage(method.Input())
				if err := protojson.Unmarshal(entry.RequestFrames[0], reqMsg); err == nil {
					activeIf = stubs.ActiveIfFromProto(reqMsg)
				}
			}
		}

		// 6. Record the stub based on type
		if !isClientStream && !isServerStream {
			// Unary success/error
			if entry.Status != 0 {
				errMsg := headers.Get("Grpc-Message")
				if decoded, err := url.QueryUnescape(errMsg); err == nil {
					errMsg = decoded
				}
				if errMsg == "" && len(entry.ResponseBody) > 0 {
					var errObj struct {
						Message string `json:"message"`
					}
					if err := json.Unmarshal(entry.ResponseBody, &errObj); err == nil {
						errMsg = errObj.Message
					}
				}
				_ = stubs.RecordErrorStub(filePath, target, activeIf, entry.Status, errMsg)
			} else {
				respMsg := dynamicpb.NewMessage(method.Output())
				if err := protojson.Unmarshal(entry.ResponseBody, respMsg); err == nil {
					_ = stubs.RecordSuccessStub(filePath, target, activeIf, respMsg)
				}
			}
		} else if isClientStream && !isServerStream {
			// Client streaming success/error
			if entry.Status != 0 {
				errMsg := headers.Get("Grpc-Message")
				if decoded, err := url.QueryUnescape(errMsg); err == nil {
					errMsg = decoded
				}
				if errMsg == "" && len(entry.ResponseBody) > 0 {
					var errObj struct {
						Message string `json:"message"`
					}
					if err := json.Unmarshal(entry.ResponseBody, &errObj); err == nil {
						errMsg = errObj.Message
					}
				}
				_ = stubs.RecordErrorStub(filePath, target, activeIf, entry.Status, errMsg)
			} else {
				respMsg := dynamicpb.NewMessage(method.Output())
				if err := protojson.Unmarshal(entry.ResponseBody, respMsg); err == nil {
					_ = stubs.RecordSuccessStub(filePath, target, activeIf, respMsg)
				}
			}
		} else {
			// Server streaming or Bidi streaming (uses RecordStreamStub)
			var items []stubs.StubFileStreamItemEntry
			for _, frame := range entry.ResponseFrames {
				var content any
				if err := json.Unmarshal(frame, &content); err == nil {
					items = append(items, stubs.StubFileStreamItemEntry{
						Content: content,
						Delay:   "100ms",
					})
				}
			}

			// If it ended with error
			if entry.Status != 0 {
				errMsg := headers.Get("Grpc-Message")
				if decoded, err := url.QueryUnescape(errMsg); err == nil {
					errMsg = decoded
				}
				items = append(items, stubs.StubFileStreamItemEntry{
					Error: &stubs.StubFileErrorEntry{
						Code:    entry.Status,
						Message: errMsg,
					},
				})
			}
			_ = stubs.RecordStreamStub(filePath, target, activeIf, items)
		}
	}
}
