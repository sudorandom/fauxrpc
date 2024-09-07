package elizav1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	fauxrpc "github.com/sudorandom/fauxrpc"
	v1 "github.com/sudorandom/fauxrpc/example/generated-faux-server/gen/connectrpc/eliza/v1"
)

type fauxElizaServiceHandler struct {
	opts fauxrpc.GenOptions
}

func NewFauxElizaServiceHandler(opts fauxrpc.GenOptions) *fauxElizaServiceHandler {
	return &fauxElizaServiceHandler{opts: opts}
}

func (h *fauxElizaServiceHandler) Say(ctx context.Context, req *connect.Request[v1.SayRequest]) (resp *connect.Response[v1.SayResponse], err error) {
	msg := &v1.SayResponse{}
	fauxrpc.SetDataOnMessage(msg, h.opts)
	return connect.NewResponse(msg), err
}

func (h *fauxElizaServiceHandler) Converse(ctx context.Context, stream *connect.BidiStream[v1.ConverseRequest, v1.ConverseResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.eliza.v1.ElizaService.Converse is not implemented"))
}

func (h *fauxElizaServiceHandler) Introduce(ctx context.Context, req *connect.Request[v1.IntroduceRequest], stream *connect.ServerStream[v1.IntroduceResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.eliza.v1.ElizaService.Introduce is not implemented"))
}
