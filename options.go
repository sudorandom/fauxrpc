package fauxrpc

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc/private/stubs"
)

type GenOptions struct {
	StubDB   stubs.StubDatabase
	MaxDepth int
	Faker    *gofakeit.Faker

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
