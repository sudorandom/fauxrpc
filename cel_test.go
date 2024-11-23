package fauxrpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/proto/gen/test/v1"
)

func TestDynamicStructNewMessage(t *testing.T) {
	t.Run("all types", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := fauxrpc.NewDynamicStruct(md, map[string]string{
			"double_value":   `1000.0+10.12`,
			"float_value":    `2000.0+10.12`,
			"int32_value":    `1+2`,
			"int64_value":    `2+2`,
			"uint32_value":   `uint(1+2)`,
			"uint64_value":   `uint(2+2)`,
			"sint32_value":   `1+2`,
			"sint64_value":   `2+2`,
			"fixed32_value":  `uint(1+2)`,
			"fixed64_value":  `uint(2+2)`,
			"sfixed32_value": `1+2`,
			"sfixed64_value": `2+2`,
			"bool_value":     `true`,
			"string_value":   `"hello"`,
			"bytes_value":    `b"ÿ"`,

			"opt_double_value":   `1000.0+10.12`,
			"opt_float_value":    `2000.0+10.12`,
			"opt_int32_value":    `1+2`,
			"opt_int64_value":    `2+2`,
			"opt_uint32_value":   `uint(1+2)`,
			"opt_uint64_value":   `uint(2+2)`,
			"opt_sint32_value":   `1+2`,
			"opt_sint64_value":   `2+2`,
			"opt_fixed32_value":  `uint(1+2)`,
			"opt_fixed64_value":  `uint(2+2)`,
			"opt_sfixed32_value": `1+2`,
			"opt_sfixed64_value": `2+2`,
			"opt_bool_value":     `true`,
			"opt_string_value":   `"hello"`,
			"opt_bytes_value":    `b"ÿ"`,
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(fauxrpc.GenOptions{})
		require.NoError(t, err)

		assert.Equal(t, 1010.12, msg.ProtoReflect().Get(md.Fields().ByTextName("double_value")).Interface())
		assert.Equal(t, float32(2010.12), msg.ProtoReflect().Get(md.Fields().ByTextName("float_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("int32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("int64_value")).Interface())
		assert.Equal(t, uint32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("uint32_value")).Interface())
		assert.Equal(t, uint64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("uint64_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("sint32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("sint64_value")).Interface())
		assert.Equal(t, uint32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("fixed32_value")).Interface())
		assert.Equal(t, uint64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("fixed64_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("sfixed32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("sfixed64_value")).Interface())
		assert.Equal(t, true, msg.ProtoReflect().Get(md.Fields().ByTextName("bool_value")).Interface())
		assert.Equal(t, "hello", msg.ProtoReflect().Get(md.Fields().ByTextName("string_value")).Interface())
		assert.Equal(t, []byte{0xc3, 0xbf}, msg.ProtoReflect().Get(md.Fields().ByTextName("bytes_value")).Interface())

		assert.Equal(t, 1010.12, msg.ProtoReflect().Get(md.Fields().ByTextName("opt_double_value")).Interface())
		assert.Equal(t, float32(2010.12), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_float_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_int32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_int64_value")).Interface())
		assert.Equal(t, uint32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_uint32_value")).Interface())
		assert.Equal(t, uint64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_uint64_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_sint32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_sint64_value")).Interface())
		assert.Equal(t, uint32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_fixed32_value")).Interface())
		assert.Equal(t, uint64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_fixed64_value")).Interface())
		assert.Equal(t, int32(3), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_sfixed32_value")).Interface())
		assert.Equal(t, int64(4), msg.ProtoReflect().Get(md.Fields().ByTextName("opt_sfixed64_value")).Interface())
		assert.Equal(t, true, msg.ProtoReflect().Get(md.Fields().ByTextName("opt_bool_value")).Interface())
		assert.Equal(t, "hello", msg.ProtoReflect().Get(md.Fields().ByTextName("opt_string_value")).Interface())
		assert.Equal(t, []byte{0xc3, 0xbf}, msg.ProtoReflect().Get(md.Fields().ByTextName("opt_bytes_value")).Interface())
	})

	t.Run("gen functions", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := fauxrpc.NewDynamicStruct(md, map[string]string{
			"double_value":   `gen_float64()`,
			"float_value":    `gen_float32()`,
			"int32_value":    `gen_int32()`,
			"int64_value":    `gen_int64()`,
			"uint32_value":   `gen_uint32()`,
			"uint64_value":   `gen_uint64()`,
			"sint32_value":   `gen_sint32()`,
			"sint64_value":   `gen_sint64()`,
			"fixed32_value":  `gen_fixed32()`,
			"fixed64_value":  `gen_fixed64()`,
			"sfixed32_value": `gen_sfixed32()`,
			"sfixed64_value": `gen_sfixed64()`,
			"bool_value":     `gen_bool()`,
			"string_value":   `gen_string()`,
			"bytes_value":    `gen_bytes()`,

			"opt_double_value":   `gen_float64()`,
			"opt_float_value":    `gen_float32()`,
			"opt_int32_value":    `gen_int32()`,
			"opt_int64_value":    `gen_int64()`,
			"opt_uint32_value":   `gen_uint32()`,
			"opt_uint64_value":   `gen_uint64()`,
			"opt_sint32_value":   `gen_sint32()`,
			"opt_sint64_value":   `gen_sint64()`,
			"opt_fixed32_value":  `gen_fixed32()`,
			"opt_fixed64_value":  `gen_fixed64()`,
			"opt_sfixed32_value": `gen_sfixed32()`,
			"opt_sfixed64_value": `gen_sfixed64()`,
			"opt_bool_value":     `gen_bool()`,
			"opt_string_value":   `gen_string()`,
			"opt_bytes_value":    `gen_bytes()`,
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(fauxrpc.GenOptions{})
		require.NoError(t, err)

		pmsg := msg.ProtoReflect()
		assertFieldIsSet(t, md, pmsg, "doubleValue")
		assertFieldIsSet(t, md, pmsg, "floatValue")
		assertFieldIsSet(t, md, pmsg, "int32Value")
		assertFieldIsSet(t, md, pmsg, "int64Value")
		assertFieldIsSet(t, md, pmsg, "uint32Value")
		assertFieldIsSet(t, md, pmsg, "uint64Value")
		assertFieldIsSet(t, md, pmsg, "sint32Value")
		assertFieldIsSet(t, md, pmsg, "sint64Value")
		assertFieldIsSet(t, md, pmsg, "fixed32Value")
		assertFieldIsSet(t, md, pmsg, "fixed64Value")
		assertFieldIsSet(t, md, pmsg, "sfixed32Value")
		assertFieldIsSet(t, md, pmsg, "sfixed64Value")
		assertFieldIsSet(t, md, pmsg, "boolValue")
		assertFieldIsSet(t, md, pmsg, "stringValue")
		assertFieldIsSet(t, md, pmsg, "bytesValue")
		assertFieldIsSet(t, md, pmsg, "optDoubleValue")
		assertFieldIsSet(t, md, pmsg, "optFloatValue")
		assertFieldIsSet(t, md, pmsg, "optInt32Value")
		assertFieldIsSet(t, md, pmsg, "optInt64Value")
		assertFieldIsSet(t, md, pmsg, "optUint32Value")
		assertFieldIsSet(t, md, pmsg, "optUint64Value")
		assertFieldIsSet(t, md, pmsg, "optSint32Value")
		assertFieldIsSet(t, md, pmsg, "optSint64Value")
		assertFieldIsSet(t, md, pmsg, "optFixed32Value")
		assertFieldIsSet(t, md, pmsg, "optFixed64Value")
		assertFieldIsSet(t, md, pmsg, "optSfixed32Value")
		assertFieldIsSet(t, md, pmsg, "optSfixed64Value")
		assertFieldIsSet(t, md, pmsg, "optBoolValue")
		assertFieldIsSet(t, md, pmsg, "optStringValue")
		assertFieldIsSet(t, md, pmsg, "optBytesValue")
		assertFieldIsSet(t, md, pmsg, "optMsgValue")
	})
}
