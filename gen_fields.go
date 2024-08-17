package fauxrpc

import (
	"strings"

	"github.com/bufbuild/protovalidate-go/resolver"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func (g *dataGenerator) getFieldValue(field protoreflect.FieldDescriptor, st state) *protoreflect.Value {
	r := resolver.DefaultResolver{}
	constraints := r.ResolveFieldConstraints(field)

	switch field.Kind() {
	case protoreflect.MessageKind:
		switch string(field.Message().FullName()) {
		case "google.protobuf.Duration":
			// TODO: hints
			return g.genGoogleDuration()
		case "google.protobuf.Timestamp":
			// TODO: hints
			return g.genGoogleTimestamp()
		case "google.protobuf.Any":
			// TODO: hints
			return nil
		case "google.protobuf.Value":
			// TODO: hints
			return g.genGoogleValue()
		default:
			nested := dynamicpb.NewMessage(field.Message())
			g.setDataOnMessage(nested, st.Inc())
			v := protoreflect.ValueOf(nested)
			return &v
		}
	case protoreflect.GroupKind:
		nested := dynamicpb.NewMessage(field.Message())
		g.setDataOnMessage(nested, st.Inc())
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.BoolKind:
		v := protoreflect.ValueOfBool(true)
		return &v
	case protoreflect.EnumKind:
		values := field.Enum().Values()
		idx := g.faker.IntRange(0, values.Len()-1)
		v := protoreflect.ValueOfEnum(protoreflect.EnumNumber(idx))
		return &v
	case protoreflect.StringKind:
		hints := StringHints{Rules: constraints.GetString_()}
		lowerName := strings.ToLower(string(field.Name()))
		switch {
		case strings.Contains(lowerName, "firstname"):
			hints.FirstName = true
		case strings.Contains(lowerName, "lastname"):
			hints.LastName = true
		case strings.Contains(lowerName, "name"):
			hints.Name = true
		case strings.Contains(lowerName, "id"):
			hints.UUID = true
		case strings.Contains(lowerName, "token"):
			hints.UUID = true
		case strings.Contains(lowerName, "url"):
			hints.URL = true
		case strings.Contains(lowerName, "version"):
			hints.Version = true
		}
		v := protoreflect.ValueOfString(GenerateString(g.faker, hints))
		return &v
	case protoreflect.BytesKind:
		hints := BytesHints{Rules: constraints.GetBytes()}
		v := protoreflect.ValueOfBytes(GenerateBytes(g.faker, hints))
		return &v
	case protoreflect.Int32Kind:
		hints := Int32Hints{Rules: constraints.GetInt32()}
		v := protoreflect.ValueOfInt32(GenerateInt32(g.faker, hints))
		return &v
	case protoreflect.Sint32Kind:
		hints := SInt32Hints{Rules: constraints.GetSint32()}
		v := protoreflect.ValueOfInt32(GenerateSInt32(g.faker, hints))
		return &v
	case protoreflect.Sfixed32Kind:
		hints := SFixedInt32Hints{Rules: constraints.GetSfixed32()}
		v := protoreflect.ValueOfInt32(GenerateSFixedInt32(g.faker, hints))
		return &v
	case protoreflect.Uint32Kind:
		hints := UInt32Hints{Rules: constraints.GetUint32()}
		v := protoreflect.ValueOfUint32(GenerateUInt32(g.faker, hints))
		return &v
	case protoreflect.Fixed32Kind:
		hints := Fixed32Hints{Rules: constraints.GetFixed32()}
		v := protoreflect.ValueOfUint32(GenerateFixed32(g.faker, hints))
		return &v
	case protoreflect.Int64Kind:
		hints := Int64Hints{Rules: constraints.GetInt64()}
		v := protoreflect.ValueOfInt64(GenerateInt64(g.faker, hints))
		return &v
	case protoreflect.Sint64Kind:
		hints := SInt64Hints{Rules: constraints.GetSint64()}
		v := protoreflect.ValueOfInt64(GenerateSInt64(g.faker, hints))
		return &v
	case protoreflect.Sfixed64Kind:
		hints := SFixed64Hints{Rules: constraints.GetSfixed64()}
		v := protoreflect.ValueOfInt64(GenerateSFixed64(g.faker, hints))
		return &v
	case protoreflect.Uint64Kind:
		hints := UInt64Hints{Rules: constraints.GetUint64()}
		v := protoreflect.ValueOfUint64(GenerateUInt64(g.faker, hints))
		return &v
	case protoreflect.Fixed64Kind:
		hints := Fixed64Hints{Rules: constraints.GetFixed64()}
		v := protoreflect.ValueOfUint64(GenerateFixed64(g.faker, hints))
		return &v
	case protoreflect.FloatKind:
		hints := Float32Hints{Rules: constraints.GetFloat()}
		v := protoreflect.ValueOfFloat32(GenerateFloat32(g.faker, hints))
		return &v
	case protoreflect.DoubleKind:
		hints := Float64Hints{Rules: constraints.GetDouble()}
		v := protoreflect.ValueOfFloat64(GenerateFloat64(g.faker, hints))
		return &v
	default:
		return nil
	}
}
