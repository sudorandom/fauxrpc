
generate:
    buf generate
    go generate ./...

run *args='': generate
    go run ./cmd/fauxrpc/ run {{ args }}

curl *args='': generate
    go run ./cmd/fauxrpc/ curl {{ args }}

test: generate
    go test ./...

lint:
    golangci-lint run ./...
