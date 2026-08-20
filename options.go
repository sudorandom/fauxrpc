package fauxrpc

import (
	"context"

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

// ExtensionResolver hands back the extensions declared for a message. It is the
// part of *protoregistry.Types that generation needs, and both
// protoregistry.GlobalTypes and a registry built from a loaded schema satisfy
// it.
type ExtensionResolver interface {
	RangeExtensionsByMessage(message protoreflect.FullName, f func(protoreflect.ExtensionType) bool)
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

	// Extensions is where generation looks for the extensions of each message
	// it fills in. It has no default: a message descriptor cannot name the
	// extensions declared against it, so without a resolver to ask, extension
	// fields are left unset.
	//
	// A registry that loaded the schema is the resolver to pass; for types
	// compiled into the binary, protoregistry.GlobalTypes works.
	Extensions ExtensionResolver

	extraFieldConstraints *validate.FieldRules
}

func (st GenOptions) GetContext() context.Context {
	if st.Context == nil {
		return context.Background()
	}
	return st.Context
}

// defaultFaker backs a GenOptions that carries no Faker of its own. It is
// created once, crypto-seeded by gofakeit.New(0), and safe for concurrent use.
//
// Creating a faker per call instead would seed each one from the wall clock,
// whose resolution is coarser than the time it takes to generate a single
// value, so consecutive values (list items, neighbouring fields) would come
// out identical.
var defaultFaker = gofakeit.New(0)

func (st GenOptions) fake() *gofakeit.Faker {
	if st.Faker == nil {
		return defaultFaker
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
