package fauxrpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestFieldValue(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
	require.NotNil(t, md)

	// Helper to get a field descriptor by name
	getField := func(fieldName string) protoreflect.FieldDescriptor {
		fd := md.Fields().ByName(protoreflect.Name(fieldName))
		require.NotNil(t, fd, "field %s not found", fieldName)
		return fd
	}

	opts := fauxrpc.GenOptions{MaxDepth: 5} // Set MaxDepth to a positive value

	t.Run("primitive types", func(t *testing.T) {
		// Test each primitive kind
		primitiveFields := []string{
			"bool_value", "string_value", "bytes_value",
			"int32_value", "sint32_value", "sfixed32_value", "uint32_value", "fixed32_value",
			"int64_value", "sint64_value", "sfixed64_value", "uint64_value", "fixed64_value",
			"float_value", "double_value",
		}

		for _, fieldName := range primitiveFields {
			t.Run(fieldName, func(t *testing.T) {
				fd := getField(fieldName)
				val := fauxrpc.FieldValue(fd, opts)
				require.NotNil(t, val)
				assert.True(t, val.IsValid())
				// Removed assert.NotZero as 0/empty is a valid value for some primitives
			})
		}
	})

	t.Run("enum type", func(t *testing.T) {
		fd := getField("enum_value")
		val := fauxrpc.FieldValue(fd, opts)
		require.NotNil(t, val)
		assert.True(t, val.IsValid())
		// Removed assert.NotZero as 0 is a valid enum value
	})

	t.Run("nested message type", func(t *testing.T) {
		fd := getField("msg_value")
		val := fauxrpc.FieldValue(fd, opts)
		require.NotNil(t, val)
		assert.True(t, val.IsValid())
		assert.True(t, val.Message().IsValid())
		// Further assertions could check if fields within the nested message are set
	})

	t.Run("well-known types", func(t *testing.T) {
		// Change message descriptor to ParameterValues for duration and timestamp
		mdParamValues := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
		require.NotNil(t, mdParamValues)

		// Helper to get a field descriptor by name from ParameterValues
		getParamField := func(fieldName string) protoreflect.FieldDescriptor {
			fd := mdParamValues.Fields().ByName(protoreflect.Name(fieldName))
			require.NotNil(t, fd, "field %s not found in ParameterValues", fieldName)
			return fd
		}

		t.Run("google.protobuf.Duration", func(t *testing.T) {
			fd := getParamField("duration")
			val := fauxrpc.FieldValue(fd, opts)
			require.NotNil(t, val)
			assert.True(t, val.IsValid())
			assert.True(t, val.Message().IsValid())
			// Check if seconds or nanos are set
			seconds := val.Message().Get(val.Message().Descriptor().Fields().ByName("seconds")).Int()
			nanos := val.Message().Get(val.Message().Descriptor().Fields().ByName("nanos")).Int()
			assert.True(t, seconds != 0 || nanos != 0)
		})

		t.Run("google.protobuf.Timestamp", func(t *testing.T) {
			fd := getParamField("timestamp")
			val := fauxrpc.FieldValue(fd, opts)
			require.NotNil(t, val)
			assert.True(t, val.IsValid())
			assert.True(t, val.Message().IsValid())
			// Check if seconds or nanos are set
			seconds := val.Message().Get(val.Message().Descriptor().Fields().ByName("seconds")).Int()
			nanos := val.Message().Get(val.Message().Descriptor().Fields().ByName("nanos")).Int()
			assert.True(t, seconds != 0 || nanos != 0)
		})

		// Removed google.protobuf.Any test as there is no corresponding field in test.proto

		// TODO: Add test for google.protobuf.Value
	})

	t.Run("MaxDepth functionality", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		msgField := md.Fields().ByName("msg_value")
		require.NotNil(t, msgField)

		// Test with MaxDepth = 0, should return nil for message fields
		optsZeroDepth := fauxrpc.GenOptions{MaxDepth: 0}
		valZeroDepth := fauxrpc.FieldValue(msgField, optsZeroDepth)
		assert.Nil(t, valZeroDepth)

		// Test with MaxDepth = 1, should generate the message but not its nested messages
		optsOneDepth := fauxrpc.GenOptions{MaxDepth: 1}
		valOneDepth := fauxrpc.FieldValue(msgField, optsOneDepth)
		require.NotNil(t, valOneDepth)
		assert.True(t, valOneDepth.IsValid())
		assert.True(t, valOneDepth.Message().IsValid())
	})
}
