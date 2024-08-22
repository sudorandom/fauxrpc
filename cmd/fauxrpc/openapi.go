package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"

	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/sudorandom/protoc-gen-connect-openapi/converter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/fauxrpc/private/registry"
)

const openapiHTML = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>FauxRPC Documentation</title>
    <script src="https://unpkg.com/@stoplight/elements@8.3.4/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements@8.3.4/styles.min.css">
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
	slices.Sort(req.FileToGenerate)
	slices.SortFunc(req.ProtoFile, func(a *descriptorpb.FileDescriptorProto, b *descriptorpb.FileDescriptorProto) int {
		return cmp.Compare(a.GetPackage(), b.GetPackage())
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
