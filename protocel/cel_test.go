package protocel_test

import (
	"context"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testv1 "github.com/sudorandom/fauxrpc/proto/gen/test/v1"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestProtocel(t *testing.T) {
	t.Run("scalars", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `
		{
			"double_value": 1000.0+10.12,
			"float_value": 2000.0+10.12,
			"int32_value": 1+2,
			"int64_value": 2+2,
			"uint32_value": uint(1+2),
			"uint64_value": uint(2+2),
			"sint32_value": 1+2,
			"sint64_value": 2+2,
			"fixed32_value": uint(1+2),
			"fixed64_value": uint(2+2),
			"sfixed32_value": 1+2,
			"sfixed64_value": 2+2,
			"bool_value": true,
			"string_value": "hello",
			"bytes_value": b"ÿ",
			"opt_double_value": 1000.0+10.12,
			"opt_float_value": 2000.0+10.12,
			"opt_int32_value": 1+2,
			"opt_int64_value": 2+2,
			"opt_uint32_value": uint(1+2),
			"opt_uint64_value": uint(2+2),
			"opt_sint32_value": 1+2,
			"opt_sint64_value": 2+2,
			"opt_fixed32_value": uint(1+2),
			"opt_fixed64_value": uint(2+2),
			"opt_sfixed32_value": 1+2,
			"opt_sfixed64_value": 2+2,
			"opt_bool_value": true,
			"opt_string_value": "hello",
			"opt_bytes_value": b"ÿ",
		}`)
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
		ds, err := protocel.New(files, md, `
		{
			"double_value": gen_float64("test.v1.AllTypes.double_value"),
			"float_value": gen_float32("test.v1.AllTypes.float_value"),
			"int32_value": gen_int32("test.v1.AllTypes.int32_value"),
			"int64_value": gen_int64("test.v1.AllTypes.int64_value"),
			"uint32_value": gen_uint32("test.v1.AllTypes.uint32_value"),
			"uint64_value": gen_uint64("test.v1.AllTypes.uint64_value"),
			"sint32_value": gen_sint32("test.v1.AllTypes.sint32_value"),
			"sint64_value": gen_sint64("test.v1.AllTypes.sint64_value"),
			"fixed32_value": gen_fixed32("test.v1.AllTypes.fixed32_value"),
			"fixed64_value": gen_fixed64("test.v1.AllTypes.fixed64_value"),
			"sfixed32_value": gen_sfixed32("test.v1.AllTypes.sfixed32_value"),
			"sfixed64_value": gen_sfixed64("test.v1.AllTypes.sfixed64_value"),
			"bool_value": gen_bool("test.v1.AllTypes.bool_value"),
			"string_value": gen_string("test.v1.AllTypes.string_value"),
			"bytes_value": gen_bytes("test.v1.AllTypes.bytes_value"),
			"opt_double_value": gen_float64("test.v1.AllTypes.opt_double_value"),
			"opt_float_value": gen_float32("test.v1.AllTypes.opt_float_value"),
			"opt_int32_value": gen_int32("test.v1.AllTypes.opt_int32_value"),
			"opt_int64_value": gen_int64("test.v1.AllTypes.opt_int64_value"),
			"opt_uint32_value": gen_uint32("test.v1.AllTypes.opt_uint32_value"),
			"opt_uint64_value": gen_uint64("test.v1.AllTypes.opt_uint64_value"),
			"opt_sint32_value": gen_sint32("test.v1.AllTypes.opt_sint32_value"),
			"opt_sint64_value": gen_sint64("test.v1.AllTypes.opt_sint64_value"),
			"opt_fixed32_value": gen_fixed32("test.v1.AllTypes.opt_fixed32_value"),
			"opt_fixed64_value": gen_fixed64("test.v1.AllTypes.opt_fixed64_value"),
			"opt_sfixed32_value": gen_sfixed32("test.v1.AllTypes.opt_sfixed32_value"),
			"opt_sfixed64_value": gen_sfixed64("test.v1.AllTypes.opt_sfixed64_value"),
			"opt_bool_value": gen_bool("test.v1.AllTypes.opt_bool_value"),
			"opt_string_value": gen_string("test.v1.AllTypes.opt_string_value"),
			"opt_bytes_value": gen_bytes("test.v1.AllTypes.opt_bytes_value"),
		}`)
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
		ds, err := protocel.New(files, md, `
		{
			"msg_value": {"string_value": "Hello World!"},
			"opt_msg_value": {"string_value": "Hello World!"},
		}`)
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
		ds, err := protocel.New(files, md, `
		{
			"msg_list": [{"string_value": "Hello World!"}],
		}`)
		require.NoError(t, err)
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
		ds, err := protocel.New(files, md, `
		{
			"string_list": ["Hello", "World!"],
			"int32_list": [1+2, 3+4],
		}`)
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
		ds, err := protocel.New(files, md, `
		{
			"string_to_string_map": {"Hello": "World!"},
			"int32_to_string_map": {1000+234: "Hello world!"},
		}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "stringToStringMap")
		assertFieldIsSet(t, md, msg.ProtoReflect(), "int32ToStringMap")

		stringToStringMap := msg.ProtoReflect().Get(md.Fields().ByTextName("string_to_string_map")).Map()
		require.Equal(t, 1, stringToStringMap.Len())
		assert.Equal(t, "World!", stringToStringMap.Get(protoreflect.MapKey(protoreflect.ValueOfString("Hello"))).Interface())

		int32ToStringMap := msg.ProtoReflect().Get(md.Fields().ByTextName("int32_to_string_map")).Map()
		require.Equal(t, 1, int32ToStringMap.Len())
		assert.Equal(t, "Hello world!", int32ToStringMap.Get(protoreflect.ValueOfInt32(int32(1234)).MapKey()).Interface())
	})

	t.Run("maps msg", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `
		{
			"msg_map": {"Hello": {"string_value": "value"}},
		}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		m := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_map")).Map()
		require.Equal(t, 1, m.Len())
		nested := m.Get(protoreflect.ValueOfString("Hello").MapKey()).Message()
		assert.Equal(t, "value", nested.Get(md.Fields().ByTextName("string_value")).Interface())
	})

	t.Run("enum", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `{"enum_value": 1}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assert.Equal(t, protoreflect.EnumNumber(1), msg.ProtoReflect().Get(md.Fields().ByTextName("enum_value")).Enum())
	})

	t.Run("enum list", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `{"enum_list": [1]}`)
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
		ds, err := protocel.New(files, md, `{"sentence": req.sentence}`)
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

	t.Run("with nested nil msg", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `{"msg_value": {"string_value": req.msg_value.string_value}}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(protocel.WithCELContext(
			context.Background(),
			&protocel.CELContext{Req: &testv1.AllTypes{}}))
		require.NoError(t, err)
		assert.True(t, proto.Equal(&testv1.AllTypes{MsgValue: &testv1.AllTypes{}}, msg))
	})

	t.Run("using req at root", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		md := elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("ConverseRequest")
		ds, err := protocel.New(files, md, `req`)
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
