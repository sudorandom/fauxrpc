package protobuf

import (
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const MaxNestedDepth = 20

type DataGenerator interface {
	SetData(msg *dynamicpb.Message)
}

type dataGenerator struct {
	faker *gofakeit.Faker
}

func NewDataGenerator() *dataGenerator {
	return &dataGenerator{faker: gofakeit.New(0)}
}

func (g *dataGenerator) SetData(msg *dynamicpb.Message) {
	// TODO: Lookup/resolve custom rules per field
	// TODO: Lookup/resolve custom rules per type, starting with well-known
	// TODO: Use known protovalidate rules as constraints
	slog.Debug("setDataOnMessage", slog.String("msg", string(msg.Descriptor().FullName())))
	defer slog.Debug("finished setDataOnMessage", slog.String("msg", string(msg.Descriptor().FullName())))
	g.setDataOnMessage(msg, 0)
}

func (g *dataGenerator) setDataOnMessage(msg *dynamicpb.Message, depth int) {
	if depth > MaxNestedDepth {
		return
	}
	desc := msg.Descriptor()
	fields := desc.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.IsList() {
			listVal := msg.NewField(field)
			itemCount := g.faker.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				if v := g.getFieldValue(field, depth+1); v != nil {
					listVal.List().Append(*v)
				} else {
					log.Printf("Unknown value %T %v", field, field.Kind())
				}
			}

			msg.Set(field, listVal)
			return
		}
		if field.IsMap() {
			mapVal := msg.NewField(field)
			itemCount := g.faker.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				v := g.getFieldValue(field.MapKey(), depth+1)
				w := g.getFieldValue(field.MapValue(), depth+1)
				if v != nil && w != nil {
					mapVal.Map().Set((*v).MapKey(), *w)
				} else {
					log.Printf("Unknown value %T %v", field, field.Kind())
				}
			}
			msg.Set(field, mapVal)
			return
		}
		if v := g.getFieldValue(field, depth+1); v != nil {
			msg.Set(field, *v)
		}
	}
}

func (g *dataGenerator) genGoogleDuration() *protoreflect.Value {
	duration := time.Duration(g.faker.IntRange(0, int(30*time.Hour*24)))
	v := protoreflect.ValueOf(durationpb.New(duration).ProtoReflect())
	return &v
}

func (g *dataGenerator) genGoogleTimestamp() *protoreflect.Value {
	v := protoreflect.ValueOf(timestamppb.New(g.faker.Date()).ProtoReflect())
	return &v
}

func (g *dataGenerator) genGoogleValue() *protoreflect.Value {
	// v := protoreflect.ValueOf(timestamppb.New(g.faker.Date()).ProtoReflect())
	options := []func() *structpb.Value{
		func() *structpb.Value { return structpb.NewNullValue() },
		func() *structpb.Value { return structpb.NewBoolValue(g.faker.Bool()) },
		func() *structpb.Value { return structpb.NewNumberValue(g.faker.Float64()) },
		func() *structpb.Value { return structpb.NewStringValue(g.faker.SentenceSimple()) },
		// TODO: structpb.NewList()
		// TODO: structpb.NewStruct()
	}
	fn := options[g.faker.IntRange(0, len(options)-1)]
	v := protoreflect.ValueOf(fn().ProtoReflect())
	return &v
}

func (g *dataGenerator) getFieldValue(field protoreflect.FieldDescriptor, depth int) *protoreflect.Value {
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
			g.setDataOnMessage(nested, depth+1)
			v := protoreflect.ValueOf(nested)
			return &v
		}
	case protoreflect.GroupKind:
		nested := dynamicpb.NewMessage(field.Message())
		g.setDataOnMessage(nested, depth+1)
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
		var v protoreflect.Value
		lowerName := strings.ToLower(string(field.Name()))
		switch {
		case strings.Contains(lowerName, "firstname"):
			v = protoreflect.ValueOfString(g.faker.FirstName())
		case strings.Contains(lowerName, "lastname"):
			v = protoreflect.ValueOfString(g.faker.LastName())
		case strings.Contains(lowerName, "name"):
			v = protoreflect.ValueOfString(g.faker.Name())
		case strings.Contains(lowerName, "id"):
			v = protoreflect.ValueOfString(g.faker.UUID())
		case strings.Contains(lowerName, "url"):
			v = protoreflect.ValueOfString(g.faker.URL())
		default:
			v = protoreflect.ValueOfString(g.faker.SentenceSimple())
		}
		return &v
	case protoreflect.Int32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Sint32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Uint32Kind:
		v := protoreflect.ValueOfUint32(g.faker.Uint32())
		return &v
	case protoreflect.Int64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Sint64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Uint64Kind:
		v := protoreflect.ValueOfUint64(g.faker.Uint64())
		return &v
	case protoreflect.Sfixed32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.Fixed32Kind:
		v := protoreflect.ValueOfInt32(g.faker.Int32())
		return &v
	case protoreflect.FloatKind:
		v := protoreflect.ValueOfFloat32(g.faker.Float32())
		return &v
	case protoreflect.Sfixed64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
		return &v
	case protoreflect.Fixed64Kind:
		v := protoreflect.ValueOfInt64(g.faker.Int64())
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
