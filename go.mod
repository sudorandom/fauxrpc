module github.com/sudorandom/fauxrpc

go 1.24.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.8-20250717185734-6c6e0d3c608e.1
	buf.build/gen/go/bufbuild/registry/connectrpc/go v1.18.1-20250819211657-a3dd0d3ea69b.1
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.8-20250819211657-a3dd0d3ea69b.1
	buf.build/gen/go/connectrpc/eliza/protocolbuffers/go v1.36.7-20230913231627-233fca715f49.1
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.18.1-20250429200738-0ee95b84c2c7.1
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.6-20250429200738-0ee95b84c2c7.1
	buf.build/go/protovalidate v0.14.0
	buf.build/go/protoyaml v0.6.0
	connectrpc.com/connect v1.18.1
	connectrpc.com/cors v0.1.0
	connectrpc.com/grpcreflect v1.3.0
	connectrpc.com/validate v0.3.0
	connectrpc.com/vanguard v0.3.0
	github.com/MadAppGang/httplog v1.3.0
	github.com/a-h/templ v0.3.924
	github.com/alecthomas/kong v1.10.0
	github.com/brianvoe/gofakeit/v7 v7.2.1
	github.com/bufbuild/protocompile v0.14.1
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/cel-go v0.26.0
	github.com/google/uuid v1.6.0
	github.com/jhump/protoreflect v1.17.0
	github.com/quic-go/quic-go v0.51.0
	github.com/rs/cors v1.11.1
	github.com/stretchr/testify v1.10.0
	github.com/sudorandom/protoc-gen-connect-openapi v0.20.3
	github.com/tailscale/hujson v0.0.0-20250226034555-ec1d1c113d33
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/net v0.43.0
	golang.org/x/sync v0.16.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.8
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
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.7.0 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/lmittmann/tint v1.1.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/pb33f/libopenapi v0.25.3 // indirect
	github.com/pb33f/ordered-map/v2 v2.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/speakeasy-api/jsonpath v0.6.2 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	go.uber.org/mock v0.5.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	github.com/a-h/templ/cmd/templ
	google.golang.org/protobuf/cmd/protoc-gen-go
)
