package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

var kindToGenerator = map[protoreflect.Kind]func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value{
	protoreflect.BoolKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfBool(Bool(fd, opts))
		return &v
	},
	protoreflect.EnumKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfEnum(Enum(fd, opts))
		return &v
	},
	protoreflect.StringKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfString(String(fd, opts))
		return &v
	},
	protoreflect.BytesKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfBytes(Bytes(fd, opts))
		return &v
	},
	protoreflect.Int32Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(Int32(fd, opts))
		return &v
	},
	protoreflect.Sint32Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(SInt32(fd, opts))
		return &v
	},
	protoreflect.Sfixed32Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt32(SFixed32(fd, opts))
		return &v
	},
	protoreflect.Uint32Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(UInt32(fd, opts))
		return &v
	},
	protoreflect.Fixed32Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfUint32(Fixed32(fd, opts))
		return &v
	},
	protoreflect.Int64Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(Int64(fd, opts))
		return &v
	},
	protoreflect.Sint64Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(SInt64(fd, opts))
		return &v
	},
	protoreflect.Sfixed64Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfInt64(SFixed64(fd, opts))
		return &v
	},
	protoreflect.Uint64Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(UInt64(fd, opts))
		return &v
	},
	protoreflect.Fixed64Kind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfUint64(Fixed64(fd, opts))
		return &v
	},
	protoreflect.FloatKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfFloat32(Float32(fd, opts))
		return &v
	},
	protoreflect.DoubleKind: func(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
		v := protoreflect.ValueOfFloat64(Float64(fd, opts))
		return &v
	},
}

func getFieldValue(fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	if opts.MaxDepth <= 0 {
		return nil
	}
	switch fd.Kind() {
	case protoreflect.MessageKind:
		switch string(fd.Message().FullName()) {
		case "google.protobuf.Duration":
			if val := GoogleDuration(fd, opts); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		case "google.protobuf.Timestamp":
			if val := GoogleTimestamp(fd, opts); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		case "google.protobuf.Any":
			return nil
		case "google.protobuf.Value":
			if val := GoogleValue(fd, opts); val != nil {
				v := protoreflect.ValueOf(val.ProtoReflect())
				return &v
			}
		}

		nested := newMessage(fd.Message())
		setDataOnMessage(nested.Interface(), opts.nested())
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.GroupKind:
		nested := newMessage(fd.Message())
		setDataOnMessage(nested.Interface(), opts.nested())
		v := protoreflect.ValueOf(nested)
		return &v
	}

	fn, ok := kindToGenerator[fd.Kind()]
	if !ok {
		return nil
	}
	return fn(fd, opts.nested())
}
