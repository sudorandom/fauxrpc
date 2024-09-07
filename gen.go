package fauxrpc

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func (st GenOptions) withExtraFieldConstraints(constraints *validate.FieldConstraints) GenOptions {
	st.extraFieldConstraints = constraints
	return st
}

func newMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}

func getFieldConstraints(fd protoreflect.FieldDescriptor, opts GenOptions) *validate.FieldConstraints {
	if constraints := getResolver().ResolveFieldConstraints(fd); constraints != nil {
		return constraints
	}
	return opts.extraFieldConstraints
}
