package fauxrpc_test

import (
	"regexp"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestString(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
	stringField := md.Fields().ByName("string_value")
	require.NotNil(t, stringField)

	t.Run("simple string", func(t *testing.T) {
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(stringField, opts)
		assert.NotEmpty(t, s)
	})

	t.Run("const rule", func(t *testing.T) {
		constVal := "fixed_string_value"
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Const: &constVal,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Equal(t, constVal, s)
	})

	t.Run("example rule", func(t *testing.T) {
		examples := []string{"example1", "example2", "example3"}
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Example: examples,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Contains(t, examples, s)
	})

	t.Run("in rule", func(t *testing.T) {
		inValues := []string{"in1", "in2", "in3"}
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					In: inValues,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Contains(t, inValues, s)
	})

	t.Run("len rule", func(t *testing.T) {
		length := uint64(10)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Len: &length,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Len(t, s, int(length))
	})

	t.Run("min_len and max_len rules", func(t *testing.T) {
		minLen := uint64(5)
		maxLen := uint64(15)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					MinLen: &minLen,
					MaxLen: &maxLen,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.GreaterOrEqual(t, len(s), int(minLen))
		assert.LessOrEqual(t, len(s), int(maxLen))
	})

	t.Run("pattern rule", func(t *testing.T) {
		pattern := "^[a-z]{5}$" // 5 lowercase letters
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Pattern: &pattern,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Regexp(t, regexp.MustCompile(pattern), s)
	})

	t.Run("well_known email", func(t *testing.T) {
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					WellKnown: &validate.StringRules_Email{
						Email: true,
					},
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		// Use a very lenient regex for validation, just checking for @ and non-empty
		assert.Regexp(t, regexp.MustCompile(`^.+@.+$`), s)
	})

	t.Run("well_known uuid", func(t *testing.T) {
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					WellKnown: &validate.StringRules_Uuid{
						Uuid: true,
					},
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Len(t, s, 36) // UUID format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
		assert.Regexp(t, regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"), s)
	})

	t.Run("field name heuristics - name", func(t *testing.T) {
		// Create a mock field descriptor directly, bypassing createFieldDescriptorWithConstraints
		nameField := &mockFieldDescriptor{
			name: "name",
			kind: protoreflect.StringKind,
		}
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(nameField, opts)
		assert.NotEmpty(t, s) // Should generate a name
	})

	t.Run("field name heuristics - id", func(t *testing.T) {
		// Create a mock field descriptor directly, bypassing createFieldDescriptorWithConstraints
		idField := &mockFieldDescriptor{
			name: "id",
			kind: protoreflect.StringKind,
		}
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(idField, opts)
		assert.Len(t, s, 36) // Should generate a UUID
	})

	t.Run("min_bytes and max_bytes rules", func(t *testing.T) {
		minBytes := uint64(5)
		maxBytes := uint64(15)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					MinBytes: &minBytes,
					MaxBytes: &maxBytes,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.GreaterOrEqual(t, len([]byte(s)), int(minBytes))
		assert.LessOrEqual(t, len([]byte(s)), int(maxBytes))
	})
}
