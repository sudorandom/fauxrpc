package main

import (
	"log/slog"
	"net/http"

	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/example/generated-faux-server/gen/connectrpc/eliza/v1/elizav1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const addr = "localhost:9000"

func main() {
	mux := http.NewServeMux()
	h := elizav1connect.NewFauxElizaServiceHandler(fauxrpc.GenOptions{})
	path, handler := elizav1connect.NewElizaServiceHandler(h)
	mux.Handle(path, handler)
	slog.Info("starting server", "addr", addr)
	http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{}))
}
