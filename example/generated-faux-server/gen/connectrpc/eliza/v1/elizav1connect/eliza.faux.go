package elizav1connect

import (
	context "context"

	connect "connectrpc.com/connect"
	fauxrpc "github.com/sudorandom/fauxrpc"
	v1 "github.com/sudorandom/fauxrpc/example/generated-faux-server/gen/connectrpc/eliza/v1"
	fauxplugin "github.com/sudorandom/fauxrpc/fauxplugin"
	"google.golang.org/protobuf/proto"
)

type fauxElizaServiceHandler struct {
	opts fauxrpc.GenOptions
}

func NewFauxElizaServiceHandler(opts fauxrpc.GenOptions) *fauxElizaServiceHandler {
	return &fauxElizaServiceHandler{opts: opts}
}

func (h *fauxElizaServiceHandler) Say(ctx context.Context, req *connect.Request[v1.SayRequest]) (resp *connect.Response[v1.SayResponse], err error) {
	var msg *v1.SayResponse
	pm := proto.Message(msg)
	return fauxplugin.UnaryHandler(ctx, h.opts, &pm)
}

func (h *fauxElizaServiceHandler) Converse(ctx context.Context, stream *connect.BidiStream[v1.ConverseRequest, v1.ConverseResponse]) error {
	return fauxplugin.BidiStreamHandler(ctx, stream, h.opts)
}

func (h *fauxElizaServiceHandler) Introduce(ctx context.Context, req *connect.Request[v1.IntroduceRequest], stream *connect.ServerStream[v1.IntroduceResponse]) error {
	return fauxplugin.ServerStreamHandler(ctx, req, stream, h.opts)
}
