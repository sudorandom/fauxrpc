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
			return g.genGoogleDuration()
		case "google.protobuf.Timestamp":
			return g.genGoogleTimestamp()
		case "google.protobuf.Any":
			return nil
		case "google.protobuf.Value":
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
		}
		v := protoreflect.ValueOfString(GenerateString(g.faker, hints))
		return &v
	case protoreflect.Int32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Sint32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Sfixed32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Uint32Kind:
		v := protoreflect.ValueOfUint32(g.faker.Uint32())
		return &v
	case protoreflect.Fixed32Kind:
		v := protoreflect.ValueOfUint32(g.faker.Uint32())
		return &v
	case protoreflect.Int64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Sint64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Sfixed64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Uint64Kind:
		v := protoreflect.ValueOfUint64(g.faker.Uint64())
		return &v
	case protoreflect.Fixed64Kind:
		v := protoreflect.ValueOfUint64(g.faker.Uint64())
		return &v
	case protoreflect.FloatKind:
		v := protoreflect.ValueOfFloat32(g.faker.Float32())
		return &v
	case protoreflect.DoubleKind:
		v := protoreflect.ValueOfFloat64(g.faker.Float64())
		return &v
	case protoreflect.BytesKind:
		v := protoreflect.ValueOfBytes([]byte(g.faker.LoremIpsumSentence(10)))
		return &v
	default:
		return nil
	}
}
