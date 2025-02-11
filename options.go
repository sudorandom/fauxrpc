package fauxrpc

import (
	"context"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type GenOptions struct {
	MaxDepth int
	Faker    *gofakeit.Faker
	Context  context.Context

	extraFieldConstraints *validate.FieldConstraints
}

func (st GenOptions) GetContext() context.Context {
	if st.Context == nil {
		return context.Background()
	}
	return st.Context
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
