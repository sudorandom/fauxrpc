package main

import (
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/mock"
	elizav1 "github.com/sudorandom/fauxrpc/example/generated-mocks/gen/connectrpc/eliza/v1"
	"github.com/sudorandom/fauxrpc/example/generated-mocks/gen/connectrpc/eliza/v1/elizav1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const addr = "localhost:9000"

func main() {
	mux := http.NewServeMux()
	h := elizav1connect.NewMockedElizaServiceHandler()
	path, handler := elizav1connect.NewElizaServiceHandler(h)
	mux.Handle(path, handler)

	h.On("Say", mock.Anything, mock.Anything).Return(connect.NewResponse(&elizav1.SayResponse{
		Sentence: "Mocked sentence here!",
	}), nil).Times(2)
	h.On("Say", mock.Anything, mock.Anything).Return(connect.NewResponse(&elizav1.SayResponse{
		Sentence: "This is the third time!",
	}), nil).Once()
	h.On("Say", mock.Anything, mock.Anything).Return(connect.NewResponse(&elizav1.SayResponse{
		Sentence: "Wow, you're calling this a lot!",
	}), nil)

	slog.Info("starting server", "addr", addr)
	http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{}))
}
