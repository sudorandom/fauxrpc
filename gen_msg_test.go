package fauxrpc_test

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/private/proto/gen/test/v1"
)

func requireFieldByName(t *testing.T, md protoreflect.MessageDescriptor, msg *dynamicpb.Message, fieldName string) protoreflect.Value {
	fd := md.Fields().ByJSONName(fieldName)
	require.NotNil(t, fd, "field %s does not exist", fieldName)
	return msg.Get(fd)
}

func assertFieldIsSet(t *testing.T, md protoreflect.MessageDescriptor, msg *dynamicpb.Message, fieldName string) {
	value := requireFieldByName(t, md, msg, fieldName)
	assert.NotNil(t, value, "field not set: %s", fieldName)
	assert.NotZero(t, value.Interface())
	assert.True(t, value.IsValid())
}

func TestGenerateMessage(t *testing.T) {
	t.Run("AllTypes", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		msg := dynamicpb.NewMessage(md)
		fauxrpc.SetDataOnMessage(msg)
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

	t.Run("ParameterValues", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
		msg := dynamicpb.NewMessage(md)
		fauxrpc.SetDataOnMessage(msg)
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

		nonZeroOneOfValues := []protoreflect.Value{}
		for _, val := range []protoreflect.Value{
			requireFieldByName(t, md, msg, "oneofDoubleValue"),
			requireFieldByName(t, md, msg, "oneofDoubleValueWrapper"),
			requireFieldByName(t, md, msg, "oneofEnumValue"),
		} {
			if !reflect.ValueOf(val.Interface()).IsZero() {
				nonZeroOneOfValues = append(nonZeroOneOfValues, val)
			}
		}
		assert.Len(t, nonZeroOneOfValues, 1, "too many values set in oneOf group: %+v", nonZeroOneOfValues)
	})
}
