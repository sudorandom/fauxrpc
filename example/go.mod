module github.com/sudorandom/fauxrpc/examples

go 1.26.3

require (
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.11-20260713175918-10d915f5b43b.1
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.36.11-20230913231627-233fca715f49.1
	buf.build/gen/go/kubernetes/cri-api/protocolbuffers/go v1.36.8-20231226185118-eb1c8c6aca91.1
	github.com/sudorandom/fauxrpc v0.16.0
	google.golang.org/protobuf v1.36.11
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20260709200747-435963d16310.1 // indirect
	buf.build/gen/go/bufbuild/registry/connectrpc/go v1.20.0-20260713175918-10d915f5b43b.1 // indirect
	buf.build/gen/go/gogo/protobuf/protocolbuffers/go v1.36.8-20220704150332-5461a3dfa9d9.1 // indirect
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.20.0-20260331211127-1730f7242d0f.1 // indirect
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.11-20260331211127-1730f7242d0f.1 // indirect
	buf.build/go/protovalidate v1.2.0 // indirect
	buf.build/go/protoyaml v0.7.0 // indirect
	cel.dev/expr v0.25.2 // indirect
	connectrpc.com/connect v1.20.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/brianvoe/gofakeit/v7 v7.15.0 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/google/cel-go v0.29.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20260709172345-9ea1abe57597 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260706201446-f0a921348800 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260727163830-6c54dddc4772 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/sudorandom/fauxrpc => ../
