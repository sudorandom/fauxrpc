package fauxrpc

import (
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GenOptions struct {
	StubDB   stubs.StubDatabase
	MaxDepth int
}

func (st GenOptions) nested() GenOptions {
	st.MaxDepth--
	return st
}

func newMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}
