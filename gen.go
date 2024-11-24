package fauxrpc

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (st GenOptions) withExtraFieldConstraints(constraints *validate.FieldConstraints) GenOptions {
	st.extraFieldConstraints = constraints
	return st
}

func getFieldConstraints(fd protoreflect.FieldDescriptor, opts GenOptions) *validate.FieldConstraints {
	if constraints := getResolver().ResolveFieldConstraints(fd); constraints != nil {
		return constraints
	}
	return opts.extraFieldConstraints
}
