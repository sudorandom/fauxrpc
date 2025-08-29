module github.com/sudorandom/fauxrpc/examples

go 1.24.6

require (
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.8-20250819211657-a3dd0d3ea69b.1
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.36.8-20230913231627-233fca715f49.1
	buf.build/gen/go/kubernetes/cri-api/protocolbuffers/go v1.36.8-20231226185118-eb1c8c6aca91.1
	github.com/sudorandom/fauxrpc v0.16.0
	google.golang.org/protobuf v1.36.8
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.8-20250717185734-6c6e0d3c608e.1 // indirect
	buf.build/gen/go/bufbuild/registry/connectrpc/go v1.18.1-20250819211657-a3dd0d3ea69b.1 // indirect
	buf.build/gen/go/gogo/protobuf/protocolbuffers/go v1.36.8-20220704150332-5461a3dfa9d9.1 // indirect
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.18.1-20250429200738-0ee95b84c2c7.1 // indirect
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.8-20250429200738-0ee95b84c2c7.1 // indirect
	buf.build/go/protovalidate v0.14.0 // indirect
	buf.build/go/protoyaml v0.6.0 // indirect
	cel.dev/expr v0.24.0 // indirect
	connectrpc.com/connect v1.18.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/brianvoe/gofakeit/v7 v7.4.0 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250826171959-ef028d996bc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sudorandom/fauxrpc => ../
