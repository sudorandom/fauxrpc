package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

var kindToGenerator = map[protoreflect.Kind]func(fd protoreflect.FieldDescriptor) *protoreflect.Value{
	protoreflect.BoolKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfBool(true)
		return &v
	},
	protoreflect.EnumKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfEnum(GenerateEnum(fd))
		return &v
	},
	protoreflect.StringKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfString(GenerateString(fd))
		return &v
	},
	protoreflect.BytesKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfBytes(GenerateBytes(fd))
		return &v
	},
	protoreflect.Int32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(GenerateInt32(fd))
		return &v
	},
	protoreflect.Sint32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(GenerateSInt32(fd))
		return &v
	},
	protoreflect.Sfixed32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(GenerateSFixed32(fd))
		return &v
	},
	protoreflect.Uint32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(GenerateUInt32(fd))
		return &v
	},
	protoreflect.Fixed32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(GenerateFixed32(fd))
		return &v
	},
	protoreflect.Int64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(GenerateInt64(fd))
		return &v
	},
	protoreflect.Sint64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(GenerateSInt64(fd))
		return &v
	},
	protoreflect.Sfixed64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(GenerateSFixed64(fd))
		return &v
	},
	protoreflect.Uint64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(GenerateUInt64(fd))
		return &v
	},
	protoreflect.Fixed64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(GenerateFixed64(fd))
		return &v
	},
	protoreflect.FloatKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfFloat32(GenerateFloat32(fd))
		return &v
	},
	protoreflect.DoubleKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfFloat64(GenerateFloat64(fd))
		return &v
	},
}

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
		}

		nested := dynamicpb.NewMessage(fd.Message())
		setDataOnMessage(nested, st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.GroupKind:
		nested := dynamicpb.NewMessage(fd.Message())
		setDataOnMessage(nested, st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	}

	fn, ok := kindToGenerator[fd.Kind()]
	if !ok {
		return nil
	}
	return fn(fd)
}
