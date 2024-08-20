package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/sudorandom/fauxrpc/private/protobuf"
	"github.com/sudorandom/protoc-gen-connect-openapi/converter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"
)

type RunCmd struct {
	Schema []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Addr   string   `short:"a" help:"Address to bind to." default:"127.0.0.1:6660"`
}

func (c *RunCmd) Run(globals *Globals) error {
	registry := protobuf.NewServiceRegistry()

	for _, schema := range c.Schema {
		if err := protobuf.AddServicesFromPath(registry, schema); err != nil {
			return err
		}
	}

	if registry.ServiceCount() == 0 {
		return errors.New("no services found in the given schemas")
	}

	// TODO: Add --no-reflection option
	// TODO: Add --no-openapi option
	// TODO: Load descriptors from stdin (assume protocol descriptors in binary format)
	// TODO: add a stub service for registering stubs

	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	registry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		vgservice := vanguard.NewServiceWithSchema(
			sd, protobuf.NewHandler(sd),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(registry.Files()))

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// OpenAPI Stuff
	resp, err := convertToOpenAPISpec(registry)
	if err != nil {
		return err
	}
	mux.Handle("GET /fauxrpc.openapi.html", singleFileHandler(openapiHTML))
	mux.Handle("GET /fauxrpc.openapi.yaml", singleFileHandler(resp.File[0].GetContent()))

	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	fmt.Printf("FauxRPC (%s)", fullVersion())
	fmt.Printf("Listening on http://%s\n", c.Addr)
	fmt.Printf("OpenAPI documentation: http://%s/fauxrpc.openapi.html\n", c.Addr)
	fmt.Println()
	fmt.Println("Example Commands:")
	fmt.Printf("$ buf curl --http2-prior-knowledge http://%s --list-methods\n", c.Addr)
	fmt.Printf("$ buf curl --http2-prior-knowledge http://%s/[METHOD_NAME]\n", c.Addr)
	return server.ListenAndServe()
}

func convertToOpenAPISpec(registry *protobuf.ServiceRegistry) (*pluginpb.CodeGeneratorResponse, error) {
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
			Title:       "FauxRPC Generated Documentation",
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

const openapiHTML = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Elements in HTML</title>
    <!-- Embed elements Elements via Web Component -->
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>

    <elements-api
      apiDescriptionUrl="/fauxrpc.openapi.yaml"
      router="hash"
      layout="sidebar"
    />

  </body>
</html>
`

type openAPIBaseInfo struct {
	Description string `yaml:"description"`
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
}
type openAPIBase struct {
	OpenAPI string          `yaml:"openapi"`
	Info    openAPIBaseInfo `yaml:"info"`
}
