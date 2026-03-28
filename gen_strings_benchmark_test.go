package fauxrpc_test

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func BenchmarkStringWithPatternAndLength(b *testing.B) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
	stringField := md.Fields().ByName("string_value")

	pattern := "^[a-z]{1,10}$"
	minLen := uint64(15) // This will always fail the length check with the above pattern
	maxLen := uint64(20)

	fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
		Type: &validate.FieldRules_String_{
			String_: &validate.StringRules{
				Pattern: &pattern,
				MinLen:  &minLen,
				MaxLen:  &maxLen,
			},
		},
	})

	opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fauxrpc.String(fd, opts)
	}
}
