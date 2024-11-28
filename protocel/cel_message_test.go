package protocel_test

import (
	"context"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testv1 "github.com/sudorandom/fauxrpc/proto/gen/test/v1"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestDynamicStructNewMessage(t *testing.T) {
	t.Run("scalars", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"double_value":   protocel.CEL(`1000.0+10.12`),
			"float_value":    protocel.CEL(`2000.0+10.12`),
			"int32_value":    protocel.CEL(`1+2`),
			"int64_value":    protocel.CEL(`2+2`),
			"uint32_value":   protocel.CEL(`uint(1+2)`),
			"uint64_value":   protocel.CEL(`uint(2+2)`),
			"sint32_value":   protocel.CEL(`1+2`),
			"sint64_value":   protocel.CEL(`2+2`),
			"fixed32_value":  protocel.CEL(`uint(1+2)`),
			"fixed64_value":  protocel.CEL(`uint(2+2)`),
			"sfixed32_value": protocel.CEL(`1+2`),
			"sfixed64_value": protocel.CEL(`2+2`),
			"bool_value":     protocel.CEL(`true`),
			"string_value":   protocel.CEL(`"hello"`),
			"bytes_value":    protocel.CEL(`b"ÿ"`),

			"opt_double_value":   protocel.CEL(`1000.0+10.12`),
			"opt_float_value":    protocel.CEL(`2000.0+10.12`),
			"opt_int32_value":    protocel.CEL(`1+2`),
			"opt_int64_value":    protocel.CEL(`2+2`),
			"opt_uint32_value":   protocel.CEL(`uint(1+2)`),
			"opt_uint64_value":   protocel.CEL(`uint(2+2)`),
			"opt_sint32_value":   protocel.CEL(`1+2`),
			"opt_sint64_value":   protocel.CEL(`2+2`),
			"opt_fixed32_value":  protocel.CEL(`uint(1+2)`),
			"opt_fixed64_value":  protocel.CEL(`uint(2+2)`),
			"opt_sfixed32_value": protocel.CEL(`1+2`),
			"opt_sfixed64_value": protocel.CEL(`2+2`),
			"opt_bool_value":     protocel.CEL(`true`),
			"opt_string_value":   protocel.CEL(`"hello"`),
			"opt_bytes_value":    protocel.CEL(`b"ÿ"`),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
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

	t.Run("scalars gen", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"double_value":   protocel.CEL(`gen_float64(field)`),
			"float_value":    protocel.CEL(`gen_float32(field)`),
			"int32_value":    protocel.CEL(`gen_int32(field)`),
			"int64_value":    protocel.CEL(`gen_int64(field)`),
			"uint32_value":   protocel.CEL(`gen_uint32(field)`),
			"uint64_value":   protocel.CEL(`gen_uint64(field)`),
			"sint32_value":   protocel.CEL(`gen_sint32(field)`),
			"sint64_value":   protocel.CEL(`gen_sint64(field)`),
			"fixed32_value":  protocel.CEL(`gen_fixed32(field)`),
			"fixed64_value":  protocel.CEL(`gen_fixed64(field)`),
			"sfixed32_value": protocel.CEL(`gen_sfixed32(field)`),
			"sfixed64_value": protocel.CEL(`gen_sfixed64(field)`),
			"bool_value":     protocel.CEL(`gen_bool(field)`),
			"string_value":   protocel.CEL(`gen_string(field)`),
			"bytes_value":    protocel.CEL(`gen_bytes(field)`),

			"opt_double_value":   protocel.CEL(`gen_float64(field)`),
			"opt_float_value":    protocel.CEL(`gen_float32(field)`),
			"opt_int32_value":    protocel.CEL(`gen_int32(field)`),
			"opt_int64_value":    protocel.CEL(`gen_int64(field)`),
			"opt_uint32_value":   protocel.CEL(`gen_uint32(field)`),
			"opt_uint64_value":   protocel.CEL(`gen_uint64(field)`),
			"opt_sint32_value":   protocel.CEL(`gen_sint32(field)`),
			"opt_sint64_value":   protocel.CEL(`gen_sint64(field)`),
			"opt_fixed32_value":  protocel.CEL(`gen_fixed32(field)`),
			"opt_fixed64_value":  protocel.CEL(`gen_fixed64(field)`),
			"opt_sfixed32_value": protocel.CEL(`gen_sfixed32(field)`),
			"opt_sfixed64_value": protocel.CEL(`gen_sfixed64(field)`),
			"opt_bool_value":     protocel.CEL(`gen_bool(field)`),
			"opt_string_value":   protocel.CEL(`gen_string(field)`),
			"opt_bytes_value":    protocel.CEL(`gen_bytes(field)`),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
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
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"msg_value": protocel.Message(map[string]protocel.Node{
				"string_value": protocel.CEL(`"Hello World!"`),
			}),
			"opt_msg_value": protocel.Message(map[string]protocel.Node{
				"string_value": protocel.CEL(`"Hello World!"`),
			}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgValue")
		nested := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_value")).Message()
		assert.Equal(t, "Hello World!", nested.Get(md.Fields().ByTextName("string_value")).Interface())
		optnested := msg.ProtoReflect().Get(md.Fields().ByTextName("opt_msg_value")).Message()
		assert.Equal(t, "Hello World!", optnested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("repeated messages", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"msg_list": protocel.Repeated([]protocel.Node{
				protocel.Message(map[string]protocel.Node{
					"string_value": protocel.CEL(`"Hello World!"`),
				}),
			}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgList")
		list := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_list")).List()
		require.Equal(t, 1, list.Len())
		nested := list.Get(0).Message()
		assert.Equal(t, "Hello World!", nested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("repeated scalars", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"string_list": protocel.Repeated([]protocel.Node{
				protocel.CEL(`"Hello"`),
				protocel.CEL(`"World!"`),
			}),
			"int32_list": protocel.Repeated([]protocel.Node{
				protocel.CEL(`1+2`),
				protocel.CEL(`3+4`),
			}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
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
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"string_to_string_map": protocel.Map(map[protocel.Node]protocel.Node{
				protocel.CEL(`"Hello!"`): protocel.CEL(`"world!"`),
			}),
			"int32_to_string_map": protocel.Map(map[protocel.Node]protocel.Node{
				protocel.CEL(`1234`): protocel.CEL(`"Hello world!"`),
			}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
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
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"msg_map": protocel.Map(map[protocel.Node]protocel.Node{
				protocel.CEL(`"Hello!"`): protocel.Message(map[string]protocel.Node{
					"string_value": protocel.CEL(`"value"`),
				}),
			}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		m := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_map")).Map()
		require.Equal(t, 1, m.Len())
		nested := m.Get(protoreflect.ValueOfString("Hello!").MapKey()).Message()
		assert.Equal(t, "value", nested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("enum", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"enum_value": protocel.CEL(`1`),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assert.Equal(t, protoreflect.EnumNumber(1), msg.ProtoReflect().Get(md.Fields().ByTextName("enum_value")).Enum())
	})

	t.Run("enum list", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"enum_list": protocel.Repeated([]protocel.Node{protocel.CEL(`1`)}),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		l := msg.ProtoReflect().Get(md.Fields().ByTextName("enum_list")).List()
		require.Equal(t, 1, l.Len())
		assert.Equal(t, protoreflect.EnumNumber(1), l.Get(0).Interface())
	})

	t.Run("with req", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		md := elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("ConverseRequest")
		ds, err := protocel.NewCELMessage(files, md, map[string]protocel.Node{
			"sentence": protocel.CEL(`req.sentence`),
		})
		require.NoError(t, err)

		msg, err := ds.NewMessage(protocel.WithCELContext(
			context.Background(),
			&protocel.CELContext{
				Req: &elizav1.ConverseRequest{
					Sentence: "hello!",
				},
			}))
		require.NoError(t, err)

		assert.Equal(t, "hello!", msg.ProtoReflect().Get(md.Fields().ByTextName("sentence")).Interface())
	})
}

func assertFieldIsSet(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) {
	value := requireFieldByName(t, md, msg, fieldName)
	assert.NotNil(t, value, "field not set: %s", fieldName)
	assert.NotZero(t, value.Interface())
	assert.True(t, value.IsValid())
}

func requireFieldByName(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) protoreflect.Value {
	fd := md.Fields().ByJSONName(fieldName)
	require.NotNil(t, fd, "field %s does not exist", fieldName)
	return msg.Get(fd)
}
