package fauxrpc_test

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func BenchmarkRepeatedUnique(b *testing.B) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
	repeatedIntField := md.Fields().ByName("int32_list")
	if repeatedIntField == nil {
		b.Fatal("field int32_list not found")
	}

	msg := dynamicpb.NewMessage(md)
	unique := true
	minItems := uint64(500) // 500 items to simulate reasonable load, O(N^2) should be noticeable
	fd := createFieldDescriptorWithConstraints(repeatedIntField, &validate.FieldRules{
		Type: &validate.FieldRules_Repeated{
			Repeated: &validate.RepeatedRules{
				Unique:   &unique,
				MinItems: &minItems,
			},
		},
	})
	opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fauxrpc.Repeated(msg, fd, opts)
	}
}
