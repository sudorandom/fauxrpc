FROM golang:latest as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build ./cmd/fauxrpc/

FROM scratch
COPY --from=builder /app/fauxrpc /fauxrpc
ENTRYPOINT ["/fauxrpc"]