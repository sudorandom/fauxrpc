module github.com/sudorandom/fauxrpc/examples/generate-messages

go 1.22.4

require (
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.34.2-20240801134127-09fbc17f7c9e.2
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.34.2-20230913231627-233fca715f49.2
	buf.build/gen/go/kubernetes/cri-api/protocolbuffers/go v1.34.2-20231226185118-eb1c8c6aca91.2
	github.com/sudorandom/fauxrpc v0.0.10
	google.golang.org/protobuf v1.34.2
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.34.2-20240717164558-a6c49f84cc0f.2 // indirect
	buf.build/gen/go/gogo/protobuf/protocolbuffers/go v1.34.2-20220704150332-5461a3dfa9d9.2 // indirect
	github.com/brianvoe/gofakeit/v7 v7.0.4 // indirect
	github.com/bufbuild/protovalidate-go v0.6.3 // indirect
)
