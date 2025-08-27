package fauxrpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestEnum(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("EnumTest")
	require.NotNil(t, md)

	opts := fauxrpc.GenOptions{MaxDepth: 5}

	getField := func(fieldName string) protoreflect.FieldDescriptor {
		fd := md.Fields().ByName(protoreflect.Name(fieldName))
		require.NotNil(t, fd, "field %s not found", fieldName)
		return fd
	}

	t.Run("in", func(t *testing.T) {
		fd := getField("enum_in")
		val := fauxrpc.Enum(fd, opts)
		assert.Contains(t, []protoreflect.EnumNumber{2, 3}, val)
	})

	t.Run("not_in", func(t *testing.T) {
		fd := getField("enum_not_in")
		val := fauxrpc.Enum(fd, opts)
		assert.NotContains(t, []protoreflect.EnumNumber{0, 1}, val)
	})

	t.Run("in_and_not_in", func(t *testing.T) {
		fd := getField("enum_in_and_not_in")
		val := fauxrpc.Enum(fd, opts)
		assert.Equal(t, protoreflect.EnumNumber(1), val)
	})

	t.Run("const", func(t *testing.T) {
		fd := getField("enum_const")
		val := fauxrpc.Enum(fd, opts)
		assert.Equal(t, protoreflect.EnumNumber(3), val)
	})

	t.Run("required", func(t *testing.T) {
		fd := getField("enum_required")
		val := fauxrpc.Enum(fd, opts)
		assert.NotEqual(t, protoreflect.EnumNumber(0), val)
	})
}
