package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/sudorandom/protoc-gen-connect-openapi/converter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1/stubsv1connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
)

type RunCmd struct {
	Schema       []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Addr         string   `short:"a" help:"Address to bind to." default:"127.0.0.1:6660"`
	NoReflection bool     `help:"Disables the server reflection service."`
	NoDocPage    bool     `help:"Disables the documentation page."`
}

func (c *RunCmd) Run(globals *Globals) error {
	theRegistry := registry.NewServiceRegistry()

	for _, schema := range c.Schema {
		if err := registry.AddServicesFromPath(theRegistry, schema); err != nil {
			return err
		}
	}

	if theRegistry.ServiceCount() == 0 {
		return errors.New("no services found in the given schemas")
	}
	// TODO: Load descriptors from stdin (assume protocol descriptors in binary format)
	// TODO: add a stub service for registering stubs

	db := stubs.NewStubDatabase()

	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	theRegistry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		vgservice := vanguard.NewServiceWithSchema(
			sd, server.NewHandler(sd, db),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)

	if !c.NoReflection {
		reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(theRegistry.Files()))
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}

	mux.Handle(stubsv1connect.NewStubsServiceHandler(stubs.NewHandler(db, theRegistry), connect.WithInterceptors(validateInterceptor)))

	// OpenAPI Stuff
	if !c.NoDocPage {
		resp, err := convertToOpenAPISpec(theRegistry)
		if err != nil {
			return err
		}
		mux.Handle("GET /fauxrpc.openapi.html", singleFileHandler(openapiHTML))
		mux.Handle("GET /fauxrpc.openapi.yaml", singleFileHandler(resp.File[0].GetContent()))
	}

	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	fmt.Printf("FauxRPC (%s)\n", fullVersion())
	fmt.Printf("Listening on http://%s\n", c.Addr)
	if !c.NoDocPage {
		fmt.Printf("OpenAPI documentation: http://%s/fauxrpc.openapi.html\n", c.Addr)
	}
	fmt.Println()
	fmt.Println("Example Commands:")
	if !c.NoReflection {
		fmt.Printf("$ buf curl --http2-prior-knowledge http://%s --list-methods\n", c.Addr)
	}
	fmt.Printf("$ buf curl --http2-prior-knowledge http://%s/[METHOD_NAME]\n", c.Addr)
	return server.ListenAndServe()
}

func convertToOpenAPISpec(registry *registry.ServiceRegistry) (*pluginpb.CodeGeneratorResponse, error) {
	req := new(plugin_go.CodeGeneratorRequest)
	files := registry.Files()
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() > 0 {
			req.FileToGenerate = append(req.FileToGenerate, string(fd.Path()))
		}
		req.ProtoFile = append(req.ProtoFile, protodesc.ToFileDescriptorProto(fd.ParentFile()))
		return true
	})
	openapiBaseFile, err := os.CreateTemp("", "base.*.openapi.yaml")
	if err != nil {
		return nil, err
	}
	defer openapiBaseFile.Close()

	descBuilder := strings.Builder{}
	descBuilder.WriteString("This is a [FauxRPC](https://fauxrpc.com/) server that is currently hosting the following services:\n")
	registry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		descBuilder.WriteString("- ")
		descBuilder.WriteString(string(sd.FullName()))
		descBuilder.WriteByte('\n')
	})
	descBuilder.WriteByte('\n')
	descBuilder.WriteString("FauxRPC is a mock server that supports gRPC, gRPC-Web, Connect and HTTP/JSON transcoding.")

	base := openAPIBase{
		OpenAPI: "3.1.0",
		Info: openAPIBaseInfo{
			Title:       "FauxRPC Documentation",
			Description: descBuilder.String(),
			Version:     strings.TrimPrefix(version, "v"),
		},
	}
	b, err := yaml.Marshal(base)
	if err != nil {
		return nil, err
	}
	if _, err := openapiBaseFile.Write(b); err != nil {
		return nil, err
	}

	req.Parameter = proto.String(fmt.Sprintf("path=all.openapi.yaml,base=%s", openapiBaseFile.Name()))

	return converter.Convert(req)
}

func singleFileHandler(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprint(w, content)
	}
}
