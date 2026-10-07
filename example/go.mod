module github.com/sudorandom/fauxrpc/examples

go 1.26.3

require (
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.11-20260713175918-10d915f5b43b.1
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.36.11-20230913231627-233fca715f49.1
	buf.build/gen/go/kubernetes/cri-api/protocolbuffers/go v1.36.8-20231226185118-eb1c8c6aca91.1
	github.com/sudorandom/fauxrpc v0.16.0
	google.golang.org/protobuf v1.36.12
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.12-20260825204119-511051f7f437.2 // indirect
	buf.build/gen/go/gogo/protobuf/protocolbuffers/go v1.36.8-20220704150332-5461a3dfa9d9.1 // indirect
	buf.build/go/protovalidate v1.4.0 // indirect
	buf.build/go/protoyaml v0.7.0 // indirect
	cel.dev/cel-go v0.32.0 // indirect
	cel.dev/expr v0.25.3 // indirect
	connectrpc.com/connect/v2 v2.0.0 // indirect
	connectrpc.com/grpcreflect/v2 v2.0.0-rc.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/brianvoe/gofakeit/v7 v7.15.0 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20260820142414-ca536658362e // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
)

replace github.com/sudorandom/fauxrpc => ../
