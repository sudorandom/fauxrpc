package fauxrpc

import (
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
			for i := 0; i < 5; i++ {
				if v := g.getFieldValue(field, depth); v != nil {
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
			for i := 0; i < 5; i++ {
				v := g.getFieldValue(field.MapKey(), depth)
				w := g.getFieldValue(field.MapValue(), depth)
				if v != nil && w != nil {
					mapVal.Map().Set((*v).MapKey(), *w)
				} else {
					log.Printf("Unknown value %T %v", field, field.Kind())
				}
			}
			msg.Set(field, mapVal)
			return
		}
		if v := g.getFieldValue(field, depth); v != nil {
			msg.Set(field, *v)
		}
	}
}

func (g *dataGenerator) getFieldValue(field protoreflect.FieldDescriptor, depth int) *protoreflect.Value {
	switch field.Kind() {
	case protoreflect.MessageKind:
		nested := dynamicpb.NewMessage(field.Message())
		g.setDataOnMessage(nested, depth+1)
		v := protoreflect.ValueOf(nested)
		return &v
	case protoreflect.GroupKind:
		nested := dynamicpb.NewMessage(field.Message())
		g.setDataOnMessage(nested, depth+1)
		v := protoreflect.ValueOf(nested)
		return &v
	default:

	}
	switch field.Kind() {
	case protoreflect.BoolKind:
		v := protoreflect.ValueOfBool(true)
		return &v
	case protoreflect.EnumKind:
		// TODO: select random enum from options
		v := protoreflect.ValueOfEnum(0)
		return &v
	case protoreflect.StringKind:
		v := protoreflect.ValueOfString(g.faker.LoremIpsumSentence(10))
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
