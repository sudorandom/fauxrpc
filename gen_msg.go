package fauxrpc

import (
	"log"

	"google.golang.org/protobuf/types/dynamicpb"
)

func (g *dataGenerator) setDataOnMessage(msg *dynamicpb.Message, st state) {
	if st.Depth > MaxNestedDepth {
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
				if v := g.getFieldValue(field, st.Inc()); v != nil {
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
				v := g.getFieldValue(field.MapKey(), st.Inc())
				w := g.getFieldValue(field.MapValue(), st.Inc())
				if v != nil && w != nil {
					mapVal.Map().Set((*v).MapKey(), *w)
				} else {
					log.Printf("Unknown value %T %v", field, field.Kind())
				}
			}
			msg.Set(field, mapVal)
			return
		}
		if v := g.getFieldValue(field, st.Inc()); v != nil {
			msg.Set(field, *v)
		}
	}
}
