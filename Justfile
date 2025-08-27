
generate:
    go generate ./...

serve *args='': generate
    go run ./cmd/fauxrpc/ run {{ args }}

curl *args='': generate
    go run ./cmd/fauxrpc/ run {{ args }}

lint:
    golangci-lint run ./...
