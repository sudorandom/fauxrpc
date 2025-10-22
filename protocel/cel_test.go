package protocel_test

import (
	"context"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
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
			"double_value": gen,
			"float_value": gen,
			"int32_value": gen,
			"int64_value": gen,
			"uint32_value": gen,
			"uint64_value": gen,
			"sint32_value": gen,
			"sint64_value": gen,
			"fixed32_value": gen,
			"fixed64_value": gen,
			"sfixed32_value": gen,
			"sfixed64_value": gen,
			"bool_value": gen,
			"string_value": gen,
			"bytes_value": gen,
			"opt_double_value": gen,
			"opt_float_value": gen,
			"opt_int32_value": gen,
			"opt_int64_value": gen,
			"opt_uint32_value": gen,
			"opt_uint64_value": gen,
			"opt_sint32_value": gen,
			"opt_sint64_value": gen,
			"opt_fixed32_value": gen,
			"opt_fixed64_value": gen,
			"opt_sfixed32_value": gen,
			"opt_sfixed64_value": gen,
			"opt_bool_value": gen,
			"opt_string_value": gen,
			"opt_bytes_value": gen,
		}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(protocel.WithCELContext(context.Background(), &protocel.CELContext{}))
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

	t.Run("scalars gen", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `gen`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(protocel.WithCELContext(context.Background(), &protocel.CELContext{}))
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

	t.Run("repeated messages with multiple elements", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `
		{
			"msg_list": [
				{"string_value": "First", "int32_value": 100},
				{"string_value": "Second", "int32_value": 200},
				{"string_value": "Third", "int32_value": 300}
			],
		}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgList")
		list := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_list")).List()
		require.Equal(t, 3, list.Len(), "Expected 3 elements in repeated field")

		// Check first element
		nested0 := list.Get(0).Message()
		assert.Equal(t, "First", nested0.Get(md.Fields().ByTextName("string_value")).Interface())
		assert.Equal(t, int32(100), nested0.Get(md.Fields().ByTextName("int32_value")).Interface())

		// Check second element
		nested1 := list.Get(1).Message()
		assert.Equal(t, "Second", nested1.Get(md.Fields().ByTextName("string_value")).Interface())
		assert.Equal(t, int32(200), nested1.Get(md.Fields().ByTextName("int32_value")).Interface())

		// Check third element
		nested2 := list.Get(2).Message()
		assert.Equal(t, "Third", nested2.Get(md.Fields().ByTextName("string_value")).Interface())
		assert.Equal(t, int32(300), nested2.Get(md.Fields().ByTextName("int32_value")).Interface())
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

	t.Run("repeated scalars with single value", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `
		{
			"string_list": "Hello",
			"int32_list": 1+2,
		}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(context.Background())
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgList")

		stringList := msg.ProtoReflect().Get(md.Fields().ByTextName("string_list")).List()
		require.Equal(t, 1, stringList.Len())
		assert.Equal(t, "Hello", stringList.Get(0).Interface())

		int32List := msg.ProtoReflect().Get(md.Fields().ByTextName("int32_list")).List()
		require.Equal(t, 1, int32List.Len())
		assert.Equal(t, int32(3), int32List.Get(0).Interface())
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
		assert.True(t, proto.Equal(testv1.AllTypes_builder{MsgValue: testv1.AllTypes_builder{StringValue: proto.String("")}.Build()}.Build(), msg))
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

	t.Run("using req at root type mismatch", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		md := elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("ConverseResponse")
		ds, err := protocel.New(files, md, `req`)
		require.NoError(t, err)

		_, err = ds.NewMessage(protocel.WithCELContext(
			context.Background(),
			&protocel.CELContext{
				Req: &elizav1.ConverseRequest{
					Sentence: "hello!",
				},
			}))
		assert.ErrorContains(t, err, "descriptor mismatch: connectrpc.eliza.v1.ConverseRequest != connectrpc.eliza.v1.ConverseRequest")
	})

	t.Run("using req msg", func(t *testing.T) {
		files := &protoregistry.Files{}
		require.NoError(t, files.RegisterFile(testv1.File_test_v1_test_proto))
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		ds, err := protocel.New(files, md, `{"msg_value": req.msg_value}`)
		require.NoError(t, err)

		msg, err := ds.NewMessage(protocel.WithCELContext(
			context.Background(),
			&protocel.CELContext{
				Req: testv1.AllTypes_builder{
					MsgValue: testv1.AllTypes_builder{
						StringValue: proto.String("Hello World!"),
					}.Build(),
				}.Build(),
			}))
		require.NoError(t, err)

		assertFieldIsSet(t, md, msg.ProtoReflect(), "msgValue")
		nested := msg.ProtoReflect().Get(md.Fields().ByTextName("msg_value")).Message()
		assert.Equal(t, "Hello World!", nested.Get(md.Fields().ByTextName("string_value")).Interface())
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
