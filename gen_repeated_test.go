package fauxrpc_test

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestRepeated(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
	repeatedStringField := md.Fields().ByName("enum_list") // Using an existing repeated field
	require.NotNil(t, repeatedStringField)

	msg := dynamicpb.NewMessage(md)

	t.Run("simple repeated field", func(t *testing.T) {
		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, repeatedStringField, opts)
		require.NotNil(t, val)
		assert.GreaterOrEqual(t, val.List().Len(), 0)
		assert.LessOrEqual(t, val.List().Len(), 4)
	})

	t.Run("min_items rule", func(t *testing.T) {
		minItems := uint64(5)
		fd := createFieldDescriptorWithConstraints(repeatedStringField, &validate.FieldRules{
			Type: &validate.FieldRules_Repeated{
				Repeated: &validate.RepeatedRules{
					MinItems: &minItems,
				},
			},
		})
		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, fd, opts)
		require.NotNil(t, val)
		assert.GreaterOrEqual(t, val.List().Len(), int(minItems))
	})

	t.Run("max_items rule", func(t *testing.T) {
		maxItems := uint64(2)
		fd := createFieldDescriptorWithConstraints(repeatedStringField, &validate.FieldRules{
			Type: &validate.FieldRules_Repeated{
				Repeated: &validate.RepeatedRules{
					MaxItems: &maxItems,
				},
			},
		})
		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, fd, opts)
		require.NotNil(t, val)
		assert.LessOrEqual(t, val.List().Len(), int(maxItems))
	})

	t.Run("min_items and max_items rules", func(t *testing.T) {
		minItems := uint64(3)
		maxItems := uint64(7)
		fd := createFieldDescriptorWithConstraints(repeatedStringField, &validate.FieldRules{
			Type: &validate.FieldRules_Repeated{
				Repeated: &validate.RepeatedRules{
					MinItems: &minItems,
					MaxItems: &maxItems,
				},
			},
		})
		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, fd, opts)
		require.NotNil(t, val)
		assert.GreaterOrEqual(t, val.List().Len(), int(minItems))
		assert.LessOrEqual(t, val.List().Len(), int(maxItems))
	})

	t.Run("unique rule for primitive types", func(t *testing.T) {
		unique := true
		minItems := uint64(5) // Ensure enough items to test uniqueness
		fd := createFieldDescriptorWithConstraints(repeatedStringField, &validate.FieldRules{
			Type: &validate.FieldRules_Repeated{
				Repeated: &validate.RepeatedRules{
					Unique:   &unique,
					MinItems: &minItems,
				},
			},
		})
		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, fd, opts)
		require.NotNil(t, val)

		// Collect all generated values and check for uniqueness
		generatedValues := make(map[any]struct{})
		for i := range val.List().Len() {
			v := val.List().Get(i).Interface()
			_, found := generatedValues[v]
			assert.False(t, found, "Duplicate value found: %v", v)
			generatedValues[v] = struct{}{}
		}
	})

	// TODO: Add test for unique rule for message types (requires mocking opts.fake() to control generated values)

	t.Run("Items rules", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		require.NotNil(t, md)
		repeatedStringField := md.Fields().ByName("string_list")
		require.NotNil(t, repeatedStringField)
		msg := dynamicpb.NewMessage(md)

		constVal := "test_value"
		fd := createFieldDescriptorWithConstraints(repeatedStringField, &validate.FieldRules{
			Type: &validate.FieldRules_Repeated{
				Repeated: &validate.RepeatedRules{
					Items: &validate.FieldRules{
						Type: &validate.FieldRules_String_{
							String_: &validate.StringRules{
								Const: &constVal,
							},
						},
					},
				},
			},
		})

		opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(0)}
		val := fauxrpc.Repeated(msg, fd, opts)
		require.NotNil(t, val)
		list := val.List()
		assert.Greater(t, list.Len(), 0)
		for i := range list.Len() {
			assert.Equal(t, constVal, list.Get(i).String())
		}
	})
}

func TestRepeatedEnum(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("RepeatedEnumTest")
	require.NotNil(t, md)

	opts := fauxrpc.GenOptions{MaxDepth: 5}

	getField := func(fieldName string) protoreflect.FieldDescriptor {
		fd := md.Fields().ByName(protoreflect.Name(fieldName))
		require.NotNil(t, fd, "field %s not found", fieldName)
		return fd
	}

	t.Run("repeated_enum_in", func(t *testing.T) {
		fd := getField("repeated_enum_in")
		msg := &testv1.RepeatedEnumTest{}
		val := fauxrpc.Repeated(msg.ProtoReflect(), fd, opts)
		require.NotNil(t, val)
		list := val.List()
		for i := range list.Len() {
			item := list.Get(i)
			assert.Contains(t, []protoreflect.EnumNumber{2, 3}, item.Enum())
		}
	})
}
