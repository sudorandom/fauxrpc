package elizav1connect

import (
	"context"
	"sync"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"connectrpc.com/connect/v2"
)

const (
	// ElizaServiceName is the fully-qualified name of the ElizaService service.
	ElizaServiceName = "connectrpc.eliza.v1.ElizaService"
)

const (
	// ElizaServiceSayProcedure is the procedure name of the ElizaService's Say RPC.
	ElizaServiceSayProcedure = "/connectrpc.eliza.v1.ElizaService/Say"
	// ElizaServiceConverseProcedure is the procedure name of the ElizaService's Converse RPC.
	ElizaServiceConverseProcedure = "/connectrpc.eliza.v1.ElizaService/Converse"
	// ElizaServiceIntroduceProcedure is the procedure name of the ElizaService's Introduce RPC.
	ElizaServiceIntroduceProcedure = "/connectrpc.eliza.v1.ElizaService/Introduce"
)

var (
	elizaServiceSaySpec = sync.OnceValue(func() connect.Spec {
		return connect.Spec{
			StreamType: connect.StreamTypeUnary,
			Procedure:  ElizaServiceSayProcedure,
		}
	})
	elizaServiceConverseSpec = sync.OnceValue(func() connect.Spec {
		return connect.Spec{
			StreamType: connect.StreamTypeBidi,
			Procedure:  ElizaServiceConverseProcedure,
		}
	})
	elizaServiceIntroduceSpec = sync.OnceValue(func() connect.Spec {
		return connect.Spec{
			StreamType: connect.StreamTypeServer,
			Procedure:  ElizaServiceIntroduceProcedure,
		}
	})
)

// ElizaServiceClient is a client for the connectrpc.eliza.v1.ElizaService service.
type ElizaServiceClient interface {
	Say(context.Context, *elizav1.SayRequest) (*elizav1.SayResponse, error)
	Converse(context.Context) (ElizaServiceConverseClientStream, error)
	Introduce(context.Context, *elizav1.IntroduceRequest) (ElizaServiceIntroduceClientStream, error)
}

// NewElizaServiceClient constructs a client for the connectrpc.eliza.v1.ElizaService service.
func NewElizaServiceClient(client *connect.Client) ElizaServiceClient {
	return &elizaServiceClient{client: client}
}

// ElizaServiceConverseClientStream is the client stream for the ElizaService's Converse RPC.
type ElizaServiceConverseClientStream struct {
	stream connect.ClientStream
}

func (s ElizaServiceConverseClientStream) Send(req *elizav1.ConverseRequest) error {
	return s.stream.Send(req)
}

func (s ElizaServiceConverseClientStream) CloseSend() error {
	return s.stream.CloseSend()
}

func (s ElizaServiceConverseClientStream) Receive() (*elizav1.ConverseResponse, error) {
	var res elizav1.ConverseResponse
	if err := s.stream.Receive(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s ElizaServiceConverseClientStream) Close() error {
	return s.stream.Close()
}

// ElizaServiceIntroduceClientStream is the client stream for the ElizaService's Introduce RPC.
type ElizaServiceIntroduceClientStream struct {
	stream connect.ClientStream
}

func (s ElizaServiceIntroduceClientStream) Receive() (*elizav1.IntroduceResponse, error) {
	var res elizav1.IntroduceResponse
	if err := s.stream.Receive(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s ElizaServiceIntroduceClientStream) Close() error {
	return s.stream.Close()
}

type elizaServiceClient struct {
	client *connect.Client
}

func (c *elizaServiceClient) Say(ctx context.Context, req *elizav1.SayRequest) (*elizav1.SayResponse, error) {
	var res elizav1.SayResponse
	if err := c.client.CallUnary(ctx, elizaServiceSaySpec(), req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *elizaServiceClient) Converse(ctx context.Context) (ElizaServiceConverseClientStream, error) {
	stream, err := c.client.CallClientStream(ctx, elizaServiceConverseSpec())
	if err != nil {
		return ElizaServiceConverseClientStream{}, err
	}
	return ElizaServiceConverseClientStream{stream: stream}, nil
}

func (c *elizaServiceClient) Introduce(ctx context.Context, req *elizav1.IntroduceRequest) (ElizaServiceIntroduceClientStream, error) {
	stream, err := c.client.CallServerStream(ctx, elizaServiceIntroduceSpec(), req)
	if err != nil {
		return ElizaServiceIntroduceClientStream{}, err
	}
	return ElizaServiceIntroduceClientStream{stream: stream}, nil
}

// ElizaServiceHandler is an implementation of the connectrpc.eliza.v1.ElizaService service.
type ElizaServiceHandler interface {
	Say(context.Context, *elizav1.SayRequest) (*elizav1.SayResponse, error)
	Converse(context.Context, ElizaServiceConverseServerStream) error
	Introduce(context.Context, *elizav1.IntroduceRequest, ElizaServiceIntroduceServerStream) error
}

type ElizaServiceConverseServerStream struct {
	stream connect.ServerStream
}

func (s ElizaServiceConverseServerStream) Receive() (*elizav1.ConverseRequest, error) {
	var req elizav1.ConverseRequest
	if err := s.stream.Receive(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s ElizaServiceConverseServerStream) Send(res *elizav1.ConverseResponse) error {
	return s.stream.Send(res)
}

type ElizaServiceIntroduceServerStream struct {
	stream connect.ServerStream
}

func (s ElizaServiceIntroduceServerStream) Send(res *elizav1.IntroduceResponse) error {
	return s.stream.Send(res)
}

// RegisterElizaServiceHandler registers svc as the connectrpc.eliza.v1.ElizaService implementation on server.
func RegisterElizaServiceHandler(server *connect.Server, svc ElizaServiceHandler) {
	adapter := elizaServiceHandler{svc: svc}
	server.Register(
		connect.Method{Spec: elizaServiceSaySpec(), Handler: adapter.say},
		connect.Method{Spec: elizaServiceConverseSpec(), Handler: adapter.converse},
		connect.Method{Spec: elizaServiceIntroduceSpec(), Handler: adapter.introduce},
	)
}

// UnimplementedElizaServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedElizaServiceHandler struct{}

func (UnimplementedElizaServiceHandler) Say(context.Context, *elizav1.SayRequest) (*elizav1.SayResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "connectrpc.eliza.v1.ElizaService.Say is not implemented")
}

func (UnimplementedElizaServiceHandler) Converse(context.Context, ElizaServiceConverseServerStream) error {
	return connect.NewError(connect.CodeUnimplemented, "connectrpc.eliza.v1.ElizaService.Converse is not implemented")
}

func (UnimplementedElizaServiceHandler) Introduce(context.Context, *elizav1.IntroduceRequest, ElizaServiceIntroduceServerStream) error {
	return connect.NewError(connect.CodeUnimplemented, "connectrpc.eliza.v1.ElizaService.Introduce is not implemented")
}

type elizaServiceHandler struct{ svc ElizaServiceHandler }

func (h elizaServiceHandler) say(ctx context.Context, _ connect.Spec, stream connect.ServerStream) error {
	var req elizav1.SayRequest
	if err := stream.Receive(&req); err != nil {
		return err
	}
	res, err := h.svc.Say(ctx, &req)
	if err != nil {
		return err
	}
	return stream.Send(res)
}

func (h elizaServiceHandler) converse(ctx context.Context, _ connect.Spec, stream connect.ServerStream) error {
	return h.svc.Converse(ctx, ElizaServiceConverseServerStream{stream: stream})
}

func (h elizaServiceHandler) introduce(ctx context.Context, _ connect.Spec, stream connect.ServerStream) error {
	var req elizav1.IntroduceRequest
	if err := stream.Receive(&req); err != nil {
		return err
	}
	return h.svc.Introduce(ctx, &req, ElizaServiceIntroduceServerStream{stream: stream})
}
