package fauxrpc

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/types/dynamicpb"
)

const MaxNestedDepth = 20

type DataGenerator interface {
	SetData(msg *dynamicpb.Message)
}

type dataGenerator struct {
	faker *gofakeit.Faker
}

func NewDataGenerator() *dataGenerator {
	return &dataGenerator{faker: gofakeit.New(0)}
}

func (g *dataGenerator) SetData(msg *dynamicpb.Message) {
	// TODO: Lookup/resolve custom rules per field
	// TODO: Lookup/resolve custom rules per type, starting with well-known
	// TODO: Use known protovalidate rules as constraints
	g.setDataOnMessage(msg, state{})
}

type state struct {
	Depth int
}

// Increment to depth and reset layer-specific values (like IsKey)
func (st state) Inc() state {
	st.Depth++
	return st
}
