package stubs

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ fauxrpc.StubFinder = (*stubFinder)(nil)

type stubFinder struct {
	db StubDatabase
}

func NewStubFinder(db StubDatabase) *stubFinder {
	return &stubFinder{
		db: db,
	}
}

func (s *stubFinder) FindStub(name protoreflect.FullName, faker *gofakeit.Faker) protoreflect.ProtoMessage {
	groups := s.db.GetStubsPrioritized(name)
	if len(groups) == 0 {
		return nil
	}
	// groups[0] is the highest priority group
	entries := groups[0]
	if len(entries) == 0 {
		return nil
	}
	// pick random
	if faker == nil {
		faker = gofakeit.GlobalFaker
	}
	idx := faker.IntRange(0, len(entries)-1)
	return entries[idx].Message
}
