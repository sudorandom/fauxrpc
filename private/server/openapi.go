package server

import (
	"strings"

	"github.com/sudorandom/protoc-gen-connect-openapi/converter"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/reflect/protoreflect"

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
      apiDescriptionUrl="/fauxrpc/openapi.yaml"
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

func convertToOpenAPISpec(registry registry.ServiceRegistry, version string) ([]byte, error) {
	descBuilder := strings.Builder{}
	descBuilder.WriteString("This is a [FauxRPC](https://fauxrpc.com/) server that is currently hosting the following services:\n")
	registry.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
		descBuilder.WriteString("- ")
		descBuilder.WriteString(string(sd.FullName()))
		descBuilder.WriteByte('\n')
		return true
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

	return converter.GenerateSingle(
		converter.WithBaseOpenAPI(b),
		converter.WithFiles(registry.Files()),
	)
}
