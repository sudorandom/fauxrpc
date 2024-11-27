package protocel_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testv1 "github.com/sudorandom/fauxrpc/proto/gen/test/v1"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestUnmarshalDynamicMessageJSON(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{}`))
		require.NoError(t, err)
		assert.NotNil(t, dmsg)
		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)
		assert.True(t, proto.Equal(&testv1.AllTypes{}, msg))
	})

	t.Run("top-level-fields", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
	"double_value": "gen_float64()",
	"float_value": "gen_float32()",
	"int32_value": "gen_int32()",
	"int64_value": "gen_int64()",
	"uint32_value": "gen_uint32()",
	"uint64_value": "gen_uint64()",
	"sint32_value": "gen_sint32()",
	"sint64_value": "gen_sint64()",
	"fixed32_value": "gen_fixed32()",
	"fixed64_value": "gen_fixed64()",
	"sfixed32_value": "gen_sfixed32()",
	"sfixed64_value": "gen_sfixed64()",
	"bool_value": "gen_bool()",
	"string_value": "gen_string()",
	"bytes_value": "gen_bytes()",

	"opt_double_value": "gen_float64()",
	"opt_float_value": "gen_float32()",
	"opt_int32_value": "gen_int32()",
	"opt_int64_value": "gen_int64()",
	"opt_uint32_value": "gen_uint32()",
	"opt_uint64_value": "gen_uint64()",
	"opt_sint32_value": "gen_sint32()",
	"opt_sint64_value": "gen_sint64()",
	"opt_fixed32_value": "gen_fixed32()",
	"opt_fixed64_value": "gen_fixed64()",
	"opt_sfixed32_value": "gen_sfixed32()",
	"opt_sfixed64_value": "gen_sfixed64()",
	"opt_bool_value": "gen_bool()",
	"opt_string_value": "gen_string()",
	"opt_bytes_value": "gen_bytes()"
}`))
		require.NoError(t, err)
		assert.NotNil(t, dmsg)
		msg, err := dmsg.NewMessage(context.Background())
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

	t.Run("nested messages", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
	"msg_value": {
		"string_value": "'Hello World!'"
	},
	"opt_msg_value": {
		"string_value": "'Hello World!'"
	}
}`))
		require.NoError(t, err)
		assert.NotNil(t, dmsg)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgValue")
		nested := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_value")).Message()
		assert.Equal(t, "Hello World!", nested.Get(md.Fields().ByTextName("string_value")).Interface())
		optnested := msg.ProtoReflect().Get(md.Fields().ByTextName("opt_msg_value")).Message()
		assert.Equal(t, "Hello World!", optnested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("repeated messages", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
			"msg_list": [{
				"string_value": "'Hello World!'"
			}]
		}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgList")
		list := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_list")).List()
		require.Equal(t, 1, list.Len())
		nested := list.Get(0).Message()
		assert.Equal(t, "Hello World!", nested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("repeated scalars", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
			"string_list": ["'Hello'", "'World!'"],
			"int32_list": ["1+2", "3+4"]
		}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgList")

		stringList := msg.ProtoReflect().Get(md.Fields().ByTextName("string_list")).List()
		require.Equal(t, 2, stringList.Len())
		assert.Equal(t, "Hello", stringList.Get(0).Interface())
		assert.Equal(t, "World!", stringList.Get(1).Interface())

		int32List := msg.ProtoReflect().Get(md.Fields().ByTextName("int32_list")).List()
		require.Equal(t, 2, int32List.Len())
		assert.Equal(t, int32(3), int32List.Get(0).Interface())
		assert.Equal(t, int32(7), int32List.Get(1).Interface())
	})

	t.Run("maps", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
			"string_to_string_map": {"'Hello!'": "'world!'"},
			"int32_to_string_map": {"1234": "'Hello world!'"}
		}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "stringToStringMap")
		assertFieldIsSet(t, md, msg.ProtoReflect(), "int32ToStringMap")

		stringToStringMap := msg.ProtoReflect().Get(md.Fields().ByTextName("string_to_string_map")).Map()
		require.Equal(t, 1, stringToStringMap.Len())
		assert.Equal(t, "world!", stringToStringMap.Get(protoreflect.MapKey(protoreflect.ValueOfString("Hello!"))).Interface())

		int32ToStringMap := msg.ProtoReflect().Get(md.Fields().ByTextName("int32_to_string_map")).Map()
		require.Equal(t, 1, int32ToStringMap.Len())
		assert.Equal(t, "Hello world!", int32ToStringMap.Get(protoreflect.ValueOfInt32(int32(1234)).MapKey()).Interface())
	})

	t.Run("maps msg", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{
			"msg_map": {"'Hello!'": {"string_value": "'value'"}}
		}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		m := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_map")).Map()
		require.Equal(t, 1, m.Len())
		nested := m.Get(protoreflect.ValueOfString("Hello!").MapKey()).Message()
		assert.Equal(t, "value", nested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("enum", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{"enum_value": "1"}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		assert.Equal(t, protoreflect.EnumNumber(1), msg.ProtoReflect().Get(md.Fields().ByTextName("enum_value")).Enum())
	})

	t.Run("enum list", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		dmsg, err := protocel.UnmarshalDynamicMessageJSON(md, []byte(`{"enum_list": ["1"]}`))
		require.NoError(t, err)

		msg, err := dmsg.NewMessage(context.Background())
		require.NoError(t, err)

		l := msg.ProtoReflect().Get(md.Fields().ByTextName("enum_list")).List()
		require.Equal(t, 1, l.Len())
		assert.Equal(t, protoreflect.EnumNumber(1), l.Get(0).Interface())
	})
}
