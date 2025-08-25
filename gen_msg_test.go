package fauxrpc_test

import (
	"fmt"
	"log"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

var AllTypes = testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")

func ExampleSetDataOnMessage() {
	msg := &elizav1.SayResponse{}
	if err := fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}); err != nil {
		log.Fatalf("error: %s", err) // handle error
	}
	b, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
	fmt.Println(string(b))
}

func ExampleNewMessage() {
	msg, _ := fauxrpc.NewMessage(elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("SayResponse"), fauxrpc.GenOptions{})
	b, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
	fmt.Println(string(b))
}

func requireFieldByName(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) protoreflect.Value {
	fd := md.Fields().ByName(protoreflect.Name(fieldName))
	require.NotNil(t, fd, "field %s does not exist", fieldName)
	return msg.Get(fd)
}

func assertFieldIsSet(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) {
	fd := md.Fields().ByName(protoreflect.Name(fieldName))
	require.NotNil(t, fd, "field %s does not exist", fieldName)
	value := msg.Get(fd)
	assert.NotNil(t, value, "field not set: %s", fieldName)
	assert.True(t, value.IsValid())
	if fd.Kind() == protoreflect.MessageKind {
		assert.True(t, value.Message().IsValid())
	} else if fd.Kind() != protoreflect.EnumKind {
		assert.NotZero(t, value.Interface())
	}
}

func TestNewMessage(t *testing.T) {
	t.Run("AllTypes - enum not_in", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		msg := dynamicpb.NewMessage(md)
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}))
		assertFieldIsSet(t, md, msg, "enum_value")
		enumValue := msg.ProtoReflect().Get(md.Fields().ByName("enum_value")).Enum()
		assert.NotEqual(t, protoreflect.EnumNumber(1), enumValue)
		assert.NotEqual(t, protoreflect.EnumNumber(2), enumValue)
	})

	t.Run("AllTypes - dynamicpb", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		msg := dynamicpb.NewMessage(md)
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}))
		assertFieldIsSet(t, md, msg, "double_value")
		assertFieldIsSet(t, md, msg, "float_value")
		assertFieldIsSet(t, md, msg, "int32_value")
		assertFieldIsSet(t, md, msg, "int64_value")
		assertFieldIsSet(t, md, msg, "uint32_value")
		assertFieldIsSet(t, md, msg, "uint64_value")
		assertFieldIsSet(t, md, msg, "sint32_value")
		assertFieldIsSet(t, md, msg, "sint64_value")
		assertFieldIsSet(t, md, msg, "fixed32_value")
		assertFieldIsSet(t, md, msg, "fixed64_value")
		assertFieldIsSet(t, md, msg, "sfixed32_value")
		assertFieldIsSet(t, md, msg, "sfixed64_value")
		assertFieldIsSet(t, md, msg, "bool_value")
		assertFieldIsSet(t, md, msg, "string_value")
		assertFieldIsSet(t, md, msg, "bytes_value")
		assertFieldIsSet(t, md, msg, "opt_double_value")
		assertFieldIsSet(t, md, msg, "opt_float_value")
		assertFieldIsSet(t, md, msg, "opt_int32_value")
		assertFieldIsSet(t, md, msg, "opt_int64_value")
		assertFieldIsSet(t, md, msg, "opt_uint32_value")
		assertFieldIsSet(t, md, msg, "opt_uint64_value")
		assertFieldIsSet(t, md, msg, "opt_sint32_value")
		assertFieldIsSet(t, md, msg, "opt_sint64_value")
		assertFieldIsSet(t, md, msg, "opt_fixed32_value")
		assertFieldIsSet(t, md, msg, "opt_fixed64_value")
		assertFieldIsSet(t, md, msg, "opt_sfixed32_value")
		assertFieldIsSet(t, md, msg, "opt_sfixed64_value")
		assertFieldIsSet(t, md, msg, "opt_bool_value")
		assertFieldIsSet(t, md, msg, "opt_string_value")
		assertFieldIsSet(t, md, msg, "opt_bytes_value")
		assertFieldIsSet(t, md, msg, "opt_msg_value")
	})

	t.Run("AllTypes - concrete", func(t *testing.T) {
		msg := &testv1.AllTypes{}
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}))
		md := msg.ProtoReflect().Descriptor()
		pmsg := msg.ProtoReflect()
		assertFieldIsSet(t, md, pmsg, "double_value")
		assertFieldIsSet(t, md, pmsg, "float_value")
		assertFieldIsSet(t, md, pmsg, "int32_value")
		assertFieldIsSet(t, md, pmsg, "int64_value")
		assertFieldIsSet(t, md, pmsg, "uint32_value")
		assertFieldIsSet(t, md, pmsg, "uint64_value")
		assertFieldIsSet(t, md, pmsg, "sint32_value")
		assertFieldIsSet(t, md, pmsg, "sint64_value")
		assertFieldIsSet(t, md, pmsg, "fixed32_value")
		assertFieldIsSet(t, md, pmsg, "fixed64_value")
		assertFieldIsSet(t, md, pmsg, "sfixed32_value")
		assertFieldIsSet(t, md, pmsg, "sfixed64_value")
		assertFieldIsSet(t, md, pmsg, "bool_value")
		assertFieldIsSet(t, md, pmsg, "string_value")
		assertFieldIsSet(t, md, pmsg, "bytes_value")
		assertFieldIsSet(t, md, pmsg, "opt_double_value")
		assertFieldIsSet(t, md, pmsg, "opt_float_value")
		assertFieldIsSet(t, md, pmsg, "opt_int32_value")
		assertFieldIsSet(t, md, pmsg, "opt_int64_value")
		assertFieldIsSet(t, md, pmsg, "opt_uint32_value")
		assertFieldIsSet(t, md, pmsg, "opt_uint64_value")
		assertFieldIsSet(t, md, pmsg, "opt_sint32_value")
		assertFieldIsSet(t, md, pmsg, "opt_sint64_value")
		assertFieldIsSet(t, md, pmsg, "opt_fixed32_value")
		assertFieldIsSet(t, md, pmsg, "opt_fixed64_value")
		assertFieldIsSet(t, md, pmsg, "opt_sfixed32_value")
		assertFieldIsSet(t, md, pmsg, "opt_sfixed64_value")
		assertFieldIsSet(t, md, pmsg, "opt_bool_value")
		assertFieldIsSet(t, md, pmsg, "opt_string_value")
		assertFieldIsSet(t, md, pmsg, "opt_bytes_value")
		assertFieldIsSet(t, md, pmsg, "opt_msg_value")
	})

	t.Run("ParameterValues - dynamicpb", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
		msg := dynamicpb.NewMessage(md)
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}))
		assertFieldIsSet(t, md, msg, "double_value")
		assertFieldIsSet(t, md, msg, "float_value")
		assertFieldIsSet(t, md, msg, "int32_value")
		assertFieldIsSet(t, md, msg, "int64_value")
		assertFieldIsSet(t, md, msg, "uint32_value")
		assertFieldIsSet(t, md, msg, "uint64_value")
		assertFieldIsSet(t, md, msg, "sint32_value")
		assertFieldIsSet(t, md, msg, "sint64_value")
		assertFieldIsSet(t, md, msg, "fixed32_value")
		assertFieldIsSet(t, md, msg, "fixed64_value")
		assertFieldIsSet(t, md, msg, "sfixed32_value")
		assertFieldIsSet(t, md, msg, "sfixed64_value")
		assertFieldIsSet(t, md, msg, "bool_value")
		assertFieldIsSet(t, md, msg, "string_value")
		assertFieldIsSet(t, md, msg, "bytes_value")
		assertFieldIsSet(t, md, msg, "timestamp")
		assertFieldIsSet(t, md, msg, "duration")
		assertFieldIsSet(t, md, msg, "bool_value_wrapper")
		assertFieldIsSet(t, md, msg, "int32_value_wrapper")
		assertFieldIsSet(t, md, msg, "int64_value_wrapper")
		assertFieldIsSet(t, md, msg, "uint32_value_wrapper")
		assertFieldIsSet(t, md, msg, "uint64_value_wrapper")
		assertFieldIsSet(t, md, msg, "float_value_wrapper")
		assertFieldIsSet(t, md, msg, "double_value_wrapper")
		assertFieldIsSet(t, md, msg, "bytes_value_wrapper")
		assertFieldIsSet(t, md, msg, "string_value_wrapper")
		assertFieldIsSet(t, md, msg, "field_mask")
		assertFieldIsSet(t, md, msg, "enum_list")
	})

	t.Run("ParameterValues - concrete", func(t *testing.T) {
		msg := &testv1.ParameterValues{}
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}))
		md := msg.ProtoReflect().Descriptor()
		pmsg := msg.ProtoReflect()
		assertFieldIsSet(t, md, pmsg, "double_value")
		assertFieldIsSet(t, md, pmsg, "float_value")
		assertFieldIsSet(t, md, pmsg, "int32_value")
		assertFieldIsSet(t, md, pmsg, "int64_value")
		assertFieldIsSet(t, md, pmsg, "uint32_value")
		assertFieldIsSet(t, md, pmsg, "uint64_value")
		assertFieldIsSet(t, md, pmsg, "sint32_value")
		assertFieldIsSet(t, md, pmsg, "sint64_value")
		assertFieldIsSet(t, md, pmsg, "fixed32_value")
		assertFieldIsSet(t, md, pmsg, "fixed64_value")
		assertFieldIsSet(t, md, pmsg, "sfixed32_value")
		assertFieldIsSet(t, md, pmsg, "sfixed64_value")
		assertFieldIsSet(t, md, pmsg, "bool_value")
		assertFieldIsSet(t, md, pmsg, "string_value")
		assertFieldIsSet(t, md, pmsg, "bytes_value")
		assertFieldIsSet(t, md, pmsg, "timestamp")
		assertFieldIsSet(t, md, pmsg, "duration")
		assertFieldIsSet(t, md, pmsg, "bool_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "int32_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "int64_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "uint32_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "uint64_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "float_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "double_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "bytes_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "string_value_wrapper")
		assertFieldIsSet(t, md, pmsg, "field_mask")
		assertFieldIsSet(t, md, pmsg, "enum_list")
	})
}
