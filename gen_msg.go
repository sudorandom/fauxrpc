package fauxrpc

import (
	"fmt"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/types/dynamicpb"
)

// SetDataOnMessage generates fake data given a *dynamicpb.Message and sets the field values.
func SetDataOnMessage(msg *dynamicpb.Message) {
	setDataOnMessage(msg, state{})
}

func setDataOnMessage(msg *dynamicpb.Message, st state) {
	if st.Depth > MaxNestedDepth {
		return
	}
	desc := msg.Descriptor()
	fields := desc.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.IsList() {
			listVal := msg.NewField(field)
			itemCount := gofakeit.GlobalFaker.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				if v := getFieldValue(field, st.Inc()); v != nil {
					listVal.List().Append(*v)
				} else {
					slog.Warn(fmt.Sprintf("Unknown list value %s %v", field.FullName(), field.Kind()))
				}
			}

			msg.Set(field, listVal)
			return
		}
		if field.IsMap() {
			mapVal := msg.NewField(field)
			itemCount := gofakeit.GlobalFaker.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				v := getFieldValue(field.MapKey(), st.Inc())
				w := getFieldValue(field.MapValue(), st.Inc())
				if v != nil && w != nil {
					mapVal.Map().Set((*v).MapKey(), *w)
				} else {
					slog.Warn(fmt.Sprintf("Unknown map k/v %s %v", field.FullName(), field.Kind()))
				}
			}
			msg.Set(field, mapVal)
			return
		}
		if v := getFieldValue(field, st.Inc()); v != nil {
			msg.Set(field, *v)
		}
	}
}
