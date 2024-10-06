package fauxplugin

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sudorandom/fauxrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func UnaryHandler[T *proto.Message](ctx context.Context, opts fauxrpc.GenOptions, msg *T) (*connect.Response[T], error) {
	fauxrpc.SetDataOnMessage(**msg, opts)
	return connect.NewResponse(msg), nil
}

func BidiStreamHandler[I proto.Message, O proto.Message](ctx context.Context, stream *connect.BidiStream[I, O], opts fauxrpc.GenOptions) error {
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for {
			_, err := stream.Receive()
			if err != nil {
				return err
			}
		}
	})
	eg.Go(func() error {
		var msg O
		fauxrpc.SetDataOnMessage(msg, opts)
		return stream.Send(&msg)
	})
	return eg.Wait()
}

func ClientStreamHandler[I proto.Message, O proto.Message](ctx context.Context, stream *connect.ClientStream[I], opts fauxrpc.GenOptions) (*connect.Response[O], error) {
	return nil, nil
}

func ServerStreamHandler[I proto.Message, O proto.Message](ctx context.Context, req *connect.Request[I], stream *connect.ServerStreamForClient[O], opts fauxrpc.GenOptions) error {
	return nil
}
