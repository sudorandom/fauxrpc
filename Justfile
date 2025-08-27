
generate:
    go generate ./...

run *args='': generate
    go run ./cmd/fauxrpc/ run {{ args }}

lint:
    golangci-lint run ./...
