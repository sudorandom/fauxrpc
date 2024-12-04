package fauxrpc

import (
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const defaultMaxDepth = 5

// NewMessage creates a new message populated with fake data given a protoreflect.MessageDescriptor
func NewMessage(md protoreflect.MessageDescriptor, opts GenOptions) (protoreflect.ProtoMessage, error) {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	opts.GetContext()
	msg := registry.NewMessage(md).Interface()
	err := setDataOnMessage(msg, opts)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// SetDataOnMessage generates fake data given a protoreflect.ProtoMessage and sets the field values.
func SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	return setDataOnMessage(msg, opts)
}

func setDataOnMessage(pm protoreflect.ProtoMessage, opts GenOptions) error {
	if opts.MaxDepth <= 0 {
		return nil
	}
	msg := pm.ProtoReflect()
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

		// pick oneOf the fields to create data for
		options := oneOf.Fields()
		idx := opts.fake().IntRange(0, options.Len()-1)
		field := options.Get(idx)
		if v := FieldValue(field, opts.nested()); v != nil {
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
			if val := Repeated(msg, field, opts); val != nil {
				msg.Set(field, *val)
			}
			continue
		}
		if field.IsMap() {
			if val := Map(msg, field, opts); val != nil {
				msg.Set(field, *val)
			}
			continue
		}
		if v := FieldValue(field, opts.nested()); v != nil {
			msg.Set(field, *v)
		}
	}
	return nil
}
