module github.com/sudorandom/fauxrpc

go 1.25.7

retract v0.15.25

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	buf.build/gen/go/bufbuild/registry/connectrpc/go v1.19.1-20251027152159-f1066ce064ca.2
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.10-20251027152159-f1066ce064ca.1
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.36.10-20230913231627-233fca715f49.1
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.19.1-20260203201457-e126be52bace.2
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.10-20260203201457-e126be52bace.1
	buf.build/go/protovalidate v1.0.0
	buf.build/go/protoyaml v0.6.0
	connectrpc.com/connect v1.19.1
	connectrpc.com/cors v0.1.0
	connectrpc.com/grpcreflect v1.3.0
	connectrpc.com/validate v0.6.0
	connectrpc.com/vanguard v0.3.0
	github.com/MadAppGang/httplog v1.3.0
	github.com/a-h/templ v0.3.960
	github.com/alecthomas/kong v1.12.1
	github.com/brianvoe/gofakeit/v7 v7.8.1
	github.com/bufbuild/protocompile v0.14.1
	github.com/go-chi/chi/v5 v5.2.3
	github.com/google/cel-go v0.26.1
	github.com/google/uuid v1.6.0
	github.com/jhump/protoreflect v1.17.0
	github.com/quic-go/quic-go v0.59.0
	github.com/rs/cors v1.11.1
	github.com/stretchr/testify v1.11.1
	github.com/sudorandom/protoc-gen-connect-openapi v0.21.3
	github.com/tailscale/hujson v0.0.0-20250226034555-ec1d1c113d33
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/net v0.47.0
	golang.org/x/sync v0.18.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.10
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2 // indirect
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cli/browser v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.7.1 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/lmittmann/tint v1.1.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/pb33f/jsonpath v0.1.2 // indirect
	github.com/pb33f/libopenapi v0.27.2 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251029180050-ab9386a59fda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	github.com/a-h/templ/cmd/templ
	google.golang.org/protobuf/cmd/protoc-gen-go
)
