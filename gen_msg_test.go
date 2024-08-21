package fauxrpc_test

import (
	"fmt"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/proto/gen/test/v1"
)

var AllTypes = testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")

func ExampleSetDataOnMessage() {
	msg := &elizav1.SayResponse{}
	fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{})
	b, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
	fmt.Println(string(b))
}

func ExampleNewMessage() {
	msg := fauxrpc.NewMessage(elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("SayResponse"), fauxrpc.GenOptions{})
	b, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
	fmt.Println(string(b))
}
func requireFieldByName(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) protoreflect.Value {
	fd := md.Fields().ByJSONName(fieldName)
	require.NotNil(t, fd, "field %s does not exist", fieldName)
	return msg.Get(fd)
}

func assertFieldIsSet(t *testing.T, md protoreflect.MessageDescriptor, msg protoreflect.Message, fieldName string) {
	value := requireFieldByName(t, md, msg, fieldName)
	assert.NotNil(t, value, "field not set: %s", fieldName)
	assert.NotZero(t, value.Interface())
	assert.True(t, value.IsValid())
}

func TestNewMessage(t *testing.T) {
	t.Run("AllTypes - dynamicpb", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		msg := dynamicpb.NewMessage(md)
		fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{})
		assertFieldIsSet(t, md, msg, "doubleValue")
		assertFieldIsSet(t, md, msg, "doubleValue")
		assertFieldIsSet(t, md, msg, "floatValue")
		assertFieldIsSet(t, md, msg, "int32Value")
		assertFieldIsSet(t, md, msg, "int64Value")
		assertFieldIsSet(t, md, msg, "uint32Value")
		assertFieldIsSet(t, md, msg, "uint64Value")
		assertFieldIsSet(t, md, msg, "sint32Value")
		assertFieldIsSet(t, md, msg, "sint64Value")
		assertFieldIsSet(t, md, msg, "fixed32Value")
		assertFieldIsSet(t, md, msg, "fixed64Value")
		assertFieldIsSet(t, md, msg, "sfixed32Value")
		assertFieldIsSet(t, md, msg, "sfixed64Value")
		assertFieldIsSet(t, md, msg, "boolValue")
		assertFieldIsSet(t, md, msg, "stringValue")
		assertFieldIsSet(t, md, msg, "bytesValue")
		assertFieldIsSet(t, md, msg, "optDoubleValue")
		assertFieldIsSet(t, md, msg, "optFloatValue")
		assertFieldIsSet(t, md, msg, "optInt32Value")
		assertFieldIsSet(t, md, msg, "optInt64Value")
		assertFieldIsSet(t, md, msg, "optUint32Value")
		assertFieldIsSet(t, md, msg, "optUint64Value")
		assertFieldIsSet(t, md, msg, "optSint32Value")
		assertFieldIsSet(t, md, msg, "optSint64Value")
		assertFieldIsSet(t, md, msg, "optFixed32Value")
		assertFieldIsSet(t, md, msg, "optFixed64Value")
		assertFieldIsSet(t, md, msg, "optSfixed32Value")
		assertFieldIsSet(t, md, msg, "optSfixed64Value")
		assertFieldIsSet(t, md, msg, "optBoolValue")
		assertFieldIsSet(t, md, msg, "optStringValue")
		assertFieldIsSet(t, md, msg, "optBytesValue")
		assertFieldIsSet(t, md, msg, "optMsgValue")
	})

	t.Run("AllTypes - concrete", func(t *testing.T) {
		msg := &testv1.AllTypes{}
		fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{})
		md := msg.ProtoReflect().Descriptor()
		pmsg := msg.ProtoReflect()
		assertFieldIsSet(t, md, pmsg, "doubleValue")
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

	t.Run("ParameterValues - dynamicpb", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
		msg := dynamicpb.NewMessage(md)
		fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{})
		assertFieldIsSet(t, md, msg, "doubleValue")
		assertFieldIsSet(t, md, msg, "floatValue")
		assertFieldIsSet(t, md, msg, "int32Value")
		assertFieldIsSet(t, md, msg, "int64Value")
		assertFieldIsSet(t, md, msg, "uint32Value")
		assertFieldIsSet(t, md, msg, "uint64Value")
		assertFieldIsSet(t, md, msg, "sint32Value")
		assertFieldIsSet(t, md, msg, "sint64Value")
		assertFieldIsSet(t, md, msg, "fixed32Value")
		assertFieldIsSet(t, md, msg, "fixed64Value")
		assertFieldIsSet(t, md, msg, "sfixed32Value")
		assertFieldIsSet(t, md, msg, "sfixed64Value")
		assertFieldIsSet(t, md, msg, "boolValue")
		assertFieldIsSet(t, md, msg, "stringValue")
		assertFieldIsSet(t, md, msg, "bytesValue")
		assertFieldIsSet(t, md, msg, "timestamp")
		assertFieldIsSet(t, md, msg, "duration")
		assertFieldIsSet(t, md, msg, "boolValueWrapper")
		assertFieldIsSet(t, md, msg, "int32ValueWrapper")
		assertFieldIsSet(t, md, msg, "int64ValueWrapper")
		assertFieldIsSet(t, md, msg, "uint32ValueWrapper")
		assertFieldIsSet(t, md, msg, "uint64ValueWrapper")
		assertFieldIsSet(t, md, msg, "floatValueWrapper")
		assertFieldIsSet(t, md, msg, "doubleValueWrapper")
		assertFieldIsSet(t, md, msg, "bytesValueWrapper")
		assertFieldIsSet(t, md, msg, "stringValueWrapper")
		assertFieldIsSet(t, md, msg, "fieldMask")
		assertFieldIsSet(t, md, msg, "enumList")
	})

	t.Run("ParameterValues - concrete", func(t *testing.T) {
		msg := &testv1.ParameterValues{}
		fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{})
		md := msg.ProtoReflect().Descriptor()
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
		assertFieldIsSet(t, md, pmsg, "timestamp")
		assertFieldIsSet(t, md, pmsg, "duration")
		assertFieldIsSet(t, md, pmsg, "boolValueWrapper")
		assertFieldIsSet(t, md, pmsg, "int32ValueWrapper")
		assertFieldIsSet(t, md, pmsg, "int64ValueWrapper")
		assertFieldIsSet(t, md, pmsg, "uint32ValueWrapper")
		assertFieldIsSet(t, md, pmsg, "uint64ValueWrapper")
		assertFieldIsSet(t, md, pmsg, "floatValueWrapper")
		assertFieldIsSet(t, md, pmsg, "doubleValueWrapper")
		assertFieldIsSet(t, md, pmsg, "bytesValueWrapper")
		assertFieldIsSet(t, md, pmsg, "stringValueWrapper")
		assertFieldIsSet(t, md, pmsg, "fieldMask")
		assertFieldIsSet(t, md, pmsg, "enumList")
	})
}
