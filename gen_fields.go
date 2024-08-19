package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

var kindToGenerator = map[protoreflect.Kind]func(fd protoreflect.FieldDescriptor) *protoreflect.Value{
	protoreflect.BoolKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfBool(true)
		return &v
	},
	protoreflect.EnumKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfEnum(Enum(fd))
		return &v
	},
	protoreflect.StringKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfString(String(fd))
		return &v
	},
	protoreflect.BytesKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfBytes(Bytes(fd))
		return &v
	},
	protoreflect.Int32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(Int32(fd))
		return &v
	},
	protoreflect.Sint32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(SInt32(fd))
		return &v
	},
	protoreflect.Sfixed32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(SFixed32(fd))
		return &v
	},
	protoreflect.Uint32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(UInt32(fd))
		return &v
	},
	protoreflect.Fixed32Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(Fixed32(fd))
		return &v
	},
	protoreflect.Int64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(Int64(fd))
		return &v
	},
	protoreflect.Sint64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(SInt64(fd))
		return &v
	},
	protoreflect.Sfixed64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(SFixed64(fd))
		return &v
	},
	protoreflect.Uint64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(UInt64(fd))
		return &v
	},
	protoreflect.Fixed64Kind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(Fixed64(fd))
		return &v
	},
	protoreflect.FloatKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfFloat32(Float32(fd))
		return &v
	},
	protoreflect.DoubleKind: func(fd protoreflect.FieldDescriptor) *protoreflect.Value {
		v := protoreflect.ValueOfFloat64(Float64(fd))
		return &v
	},
}

func getFieldValue(fd protoreflect.FieldDescriptor, st state) *protoreflect.Value {
	switch fd.Kind() {
	case protoreflect.MessageKind:
		switch string(fd.Message().FullName()) {
		case "google.protobuf.Duration":
			if val := GoogleDuration(fd); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		case "google.protobuf.Timestamp":
			if val := GoogleTimestamp(fd); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		case "google.protobuf.Any":
			return nil
		case "google.protobuf.Value":
			if val := GoogleValue(fd, st); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		}

		var nested protoreflect.Message
		mt, err := protoregistry.GlobalTypes.FindMessageByName(fd.Message().FullName())
		if err != nil {
			nested = dynamicpb.NewMessageType(fd.Message()).New()
		} else {
			nested = mt.New()
		}
		setDataOnMessage(nested.Interface(), st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.GroupKind:
		var nested protoreflect.Message
		mt, err := protoregistry.GlobalTypes.FindMessageByName(fd.Message().FullName())
		if err != nil {
			nested = dynamicpb.NewMessageType(fd.Message()).New()
		} else {
			nested = mt.New()
		}
		setDataOnMessage(nested.Interface(), st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	}

	fn, ok := kindToGenerator[fd.Kind()]
	if !ok {
		return nil
	}
	return fn(fd)
}
