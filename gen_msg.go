package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

const defaultMaxDepth = 20

// NewMessage creates a new message populated with fake data given a protoreflect.MessageDescriptor
func NewMessage(md protoreflect.MessageDescriptor, opts GenOptions) protoreflect.ProtoMessage {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	msg := newMessage(md).Interface()
	setDataOnMessage(msg, opts)
	return msg
}

// SetDataOnMessage generates fake data given a protoreflect.ProtoMessage and sets the field values.
func SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	setDataOnMessage(msg, opts)
}

func setDataOnMessage(pm protoreflect.ProtoMessage, opts GenOptions) {
	if opts.MaxDepth <= 0 {
		return
	}
	msg := pm.ProtoReflect()
	desc := msg.Descriptor()

	if opts.StubDB != nil {
		stubs := opts.StubDB.GetStubs(desc.FullName())
		if len(stubs) > 0 {
			idx := opts.fake().IntRange(0, len(stubs)-1)
			other := stubs[idx]
			fields := desc.Fields()
			for i := 0; i < fields.Len(); i++ {
				field := fields.Get(i)
				if other.ProtoReflect().Has(field) {
					msg.Set(field, other.ProtoReflect().Get(field))
				}
			}
			return
		}
	}

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
		if v := getFieldValue(field, opts.nested()); v != nil {
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
			return
		}
		if field.IsMap() {
			if val := Map(msg, field, opts); val != nil {
				msg.Set(field, *val)
			}
			return
		}
		if v := getFieldValue(field, opts.nested()); v != nil {
			msg.Set(field, *v)
		}
	}
}
