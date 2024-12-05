package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

func repeatedSimple(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	listVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(0, 4)
	for i := 0; i < itemCount; i++ {
		if v := FieldValue(fd, opts.nested()); v != nil {
			listVal.List().Append(*v)
		}
	}
	return &listVal
}

// Repeated returns a fake repeated value given a field descriptor.
func Repeated(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return repeatedSimple(msg, fd, opts)
	}
	rules := constraints.GetEnum()
	if rules == nil {
		return repeatedSimple(msg, fd, opts)
	}
	min, max := uint64(0), uint64(4)
	if constraints.GetRepeated().MinItems != nil {
		min = constraints.GetRepeated().GetMinItems()
	}
	if constraints.GetRepeated().MaxItems != nil {
		max = constraints.GetRepeated().GetMaxItems()
	}

	listVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(int(min), int(max))
	for i := 0; i < itemCount; i++ {
		if v := FieldValue(fd, opts.nested().withExtraFieldConstraints(constraints.GetRepeated().Items)); v != nil {
			listVal.List().Append(*v)
		}
	}
	return &listVal
}
