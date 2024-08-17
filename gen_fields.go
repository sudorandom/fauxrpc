package fauxrpc

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func getFieldValue(fd protoreflect.FieldDescriptor, st state) *protoreflect.Value {
	switch fd.Kind() {
	case protoreflect.MessageKind:
		switch string(fd.Message().FullName()) {
		case "google.protobuf.Duration":
			// TODO: hints
			return genGoogleDuration()
		case "google.protobuf.Timestamp":
			// TODO: hints
			return genGoogleTimestamp()
		case "google.protobuf.Any":
			// TODO: hints
			return nil
		case "google.protobuf.Value":
			// TODO: hints
			return genGoogleValue()
		default:
			nested := dynamicpb.NewMessage(fd.Message())
			setDataOnMessage(nested, st.Inc())
			v := protoreflect.ValueOf(nested)
			return &v
		}
	case protoreflect.GroupKind:
		nested := dynamicpb.NewMessage(fd.Message())
		setDataOnMessage(nested, st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.BoolKind:
		v := protoreflect.ValueOfBool(true)
		return &v
	case protoreflect.EnumKind:
		values := fd.Enum().Values()
		idx := gofakeit.IntRange(0, values.Len()-1)
		v := protoreflect.ValueOfEnum(protoreflect.EnumNumber(idx))
		return &v
	case protoreflect.StringKind:
		v := protoreflect.ValueOfString(GenerateString(fd))
		return &v
	case protoreflect.BytesKind:
		v := protoreflect.ValueOfBytes(GenerateBytes(fd))
		return &v
	case protoreflect.Int32Kind:
		v := protoreflect.ValueOfInt32(GenerateInt32(fd))
		return &v
	case protoreflect.Sint32Kind:
		v := protoreflect.ValueOfInt32(GenerateSInt32(fd))
		return &v
	case protoreflect.Sfixed32Kind:
		v := protoreflect.ValueOfInt32(GenerateSFixedInt32(fd))
		return &v
	case protoreflect.Uint32Kind:
		v := protoreflect.ValueOfUint32(GenerateUInt32(fd))
		return &v
	case protoreflect.Fixed32Kind:
		v := protoreflect.ValueOfUint32(GenerateFixed32(fd))
		return &v
	case protoreflect.Int64Kind:
		v := protoreflect.ValueOfInt64(GenerateInt64(fd))
		return &v
	case protoreflect.Sint64Kind:
		v := protoreflect.ValueOfInt64(GenerateSInt64(fd))
		return &v
	case protoreflect.Sfixed64Kind:
		v := protoreflect.ValueOfInt64(GenerateSFixed64(fd))
		return &v
	case protoreflect.Uint64Kind:
		v := protoreflect.ValueOfUint64(GenerateUInt64(fd))
		return &v
	case protoreflect.Fixed64Kind:
		v := protoreflect.ValueOfUint64(GenerateFixed64(fd))
		return &v
	case protoreflect.FloatKind:
		v := protoreflect.ValueOfFloat32(GenerateFloat32(fd))
		return &v
	case protoreflect.DoubleKind:
		v := protoreflect.ValueOfFloat64(GenerateFloat64(fd))
		return &v
	default:
		return nil
	}
}
