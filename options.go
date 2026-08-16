package fauxrpc

import (
	"context"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type StubEntry interface {
	GetName() protoreflect.FullName
	GetID() string
}

type StubFinder interface {
	FindStub(name protoreflect.FullName, faker *gofakeit.Faker) protoreflect.ProtoMessage
}

type FieldGenOptions struct {
	Message *validate.FieldRules
}

type GenOptions struct {
	MaxDepth     int
	Faker        *gofakeit.Faker
	Context      context.Context
	StubRecorder func(StubEntry)
	StubFinder   StubFinder

	// ViolateRules is the probability, from 0.0 to 1.0, that any single
	// protovalidate rule is broken rather than satisfied. Every rule on every
	// field is rolled independently, so a field carrying five rules is far more
	// likely to end up invalid than a field carrying one. 0 (the default) never
	// violates and 1 breaks every rule it can.
	//
	// A field can only hold one value, so when several of its rules come up at
	// once, one of them is picked at random to be the one that breaks.
	//
	// Rules that no value can violate on its own (message-level rules and CEL
	// expressions) are always satisfied, so a message is not guaranteed to fail
	// validation even at 1.0.
	ViolateRules float64

	extraFieldConstraints *validate.FieldRules
}

func (st GenOptions) GetContext() context.Context {
	if st.Context == nil {
		return context.Background()
	}
	return st.Context
}

func (st GenOptions) fake() *gofakeit.Faker {
	if st.Faker == nil {
		return gofakeit.New(uint64(time.Now().UnixNano()))
	}
	return st.Faker
}

// shouldViolate rolls the ViolateRules dice for a single rule.
func (st GenOptions) shouldViolate() bool {
	if st.ViolateRules <= 0 {
		return false
	}
	if st.ViolateRules >= 1 {
		return true
	}
	return st.fake().Float64() < st.ViolateRules
}

func (st GenOptions) nested() GenOptions {
	st.MaxDepth--
	st.extraFieldConstraints = nil
	return st
}

func (st GenOptions) WithExtraFieldConstraints(rules *validate.FieldRules) GenOptions {
	st.extraFieldConstraints = rules
	return st
}
