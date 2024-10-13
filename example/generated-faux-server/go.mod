module github.com/sudorandom/fauxrpc/example/generated-faux-server

go 1.23.2

require (
	connectrpc.com/connect v1.17.0
	github.com/sudorandom/fauxrpc v0.0.17
	golang.org/x/net v0.30.0
	google.golang.org/protobuf v1.35.1
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.35.1-20240920164238-5a7b106cbb87.1 // indirect
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.16.2-20240809200651-8507e5a24938.1 // indirect
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.34.1-20240809200651-8507e5a24938.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/brianvoe/gofakeit/v7 v7.0.4 // indirect
	github.com/bufbuild/protocompile v0.14.0 // indirect
	github.com/bufbuild/protovalidate-go v0.7.2 // indirect
	github.com/bufbuild/protoyaml-go v0.1.11 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.21.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/exp v0.0.0-20241004190924-225e2abe05e6 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240930140551-af27646dc61f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sudorandom/fauxrpc => ../../
