package fauxrpc_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func BenchmarkEnumFunction(b *testing.B) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("EnumTest")
	fd := md.Fields().ByName(protoreflect.Name("enum_in"))
	opts := fauxrpc.GenOptions{MaxDepth: 3, Faker: gofakeit.New(0)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fauxrpc.Enum(fd, opts)
	}
}
