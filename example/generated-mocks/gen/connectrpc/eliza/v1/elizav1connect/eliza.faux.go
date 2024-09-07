package elizav1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	mock "github.com/stretchr/testify/mock"
	v1 "github.com/sudorandom/fauxrpc/example/generated-mocks/gen/connectrpc/eliza/v1"
)

type mockedElizaServiceHandler struct {
	mock.Mock
}

func NewMockedElizaServiceHandler() *mockedElizaServiceHandler {
	h := new(mockedElizaServiceHandler)
	return h
}

func (h *mockedElizaServiceHandler) Say(ctx context.Context, req *connect.Request[v1.SayRequest]) (resp *connect.Response[v1.SayResponse], err error) {
	args := h.Called(ctx, req)
	resp, _ = args.Get(0).(*connect.Response[v1.SayResponse])
	err, _ = args.Get(1).(error)
	return resp, err
}

func (h *mockedElizaServiceHandler) Converse(ctx context.Context, stream *connect.BidiStream[v1.ConverseRequest, v1.ConverseResponse]) error {
	h.Called(ctx, stream)
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.eliza.v1.ElizaService.Converse is not implemented"))
}

func (h *mockedElizaServiceHandler) Introduce(ctx context.Context, req *connect.Request[v1.IntroduceRequest], stream *connect.ServerStream[v1.IntroduceResponse]) error {
	h.Called(ctx, stream)
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.eliza.v1.ElizaService.Introduce is not implemented"))
}
