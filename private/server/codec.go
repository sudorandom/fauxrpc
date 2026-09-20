package server

import (
	"context"
	"io"

	"connectrpc.com/connect/v2/connectproto"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// requestMessage is the sentinel value method handlers pass to
// ServerStream.Receive. The codecs below recognize it and decode the request
// with the most efficient parser available, storing the result in msg.
type requestMessage struct {
	desc protoreflect.MessageDescriptor
	// data keeps the request bytes reachable for zero-copy parsers that
	// alias them.
	data []byte
	msg  releasableMessage
}

type plainMessage struct {
	proto.Message
}

func (plainMessage) Release() {}

func newBinaryCodec(resolver connectproto.TypeResolver) *binaryCodec {
	return &binaryCodec{BinaryCodec: connectproto.NewBinaryCodec(connectproto.WithTypeResolver(resolver))}
}

type binaryCodec struct {
	*connectproto.BinaryCodec
}

func (c *binaryCodec) UnmarshalRead(ctx context.Context, src io.Reader, msg any) error {
	req, ok := msg.(*requestMessage)
	if !ok {
		return c.BinaryCodec.UnmarshalRead(ctx, src, msg)
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	decoded, err := unmarshalRequest(req.desc, data)
	if err != nil {
		return err
	}
	req.data = data
	req.msg = decoded
	return nil
}

func newJSONCodec(resolver connectproto.TypeResolver) *jsonCodec {
	return &jsonCodec{JSONCodec: connectproto.NewJSONCodec(connectproto.WithTypeResolver(resolver))}
}

type jsonCodec struct {
	*connectproto.JSONCodec
}

func (c *jsonCodec) UnmarshalRead(ctx context.Context, src io.Reader, msg any) error {
	req, ok := msg.(*requestMessage)
	if !ok {
		return c.JSONCodec.UnmarshalRead(ctx, src, msg)
	}
	decoded := registry.NewMessage(req.desc).Interface()
	if err := c.JSONCodec.UnmarshalRead(ctx, src, decoded); err != nil {
		return err
	}
	req.msg = plainMessage{Message: decoded}
	return nil
}
