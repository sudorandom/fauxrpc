package fauxrpc

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GenOptions struct {
	StubDB           stubs.StubDatabase
	MaxDepth         int
	Faker            *gofakeit.Faker
	MethodDescriptor protoreflect.MethodDescriptor
	Input            proto.Message

	extraFieldConstraints *validate.FieldConstraints
}

func (st GenOptions) fake() *gofakeit.Faker {
	if st.Faker == nil {
		return gofakeit.GlobalFaker
	}
	return st.Faker
}

func (st GenOptions) nested() GenOptions {
	st.MaxDepth--
	st.extraFieldConstraints = nil
	return st
}
