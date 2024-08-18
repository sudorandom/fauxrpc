package fauxrpc

import (
	"fmt"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	oneOfFields := map[protoreflect.FullName]struct{}{}
	oneOfs := desc.Oneofs()
	// gather one-of fields
	for i := 0; i < oneOfs.Len(); i++ {
		oneOf := oneOfs.Get(i)
		fields := oneOf.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			oneOfFields[field.FullName()] = struct{}{}
		}
		options := oneOf.Fields()
		idx := gofakeit.IntRange(0, options.Len()-1)
		field := options.Get(idx)
		if v := getFieldValue(field, st.Inc()); v != nil {
			msg.Set(field, *v)
		}
	}

	fields := desc.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if _, ok := oneOfFields[field.FullName()]; ok {
			continue
		}
		if field.IsList() {
			listVal := msg.NewField(field)
			itemCount := gofakeit.IntRange(0, 4)
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
			itemCount := gofakeit.IntRange(0, 4)
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
