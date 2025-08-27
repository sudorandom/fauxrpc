package fauxrpc

import (
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func repeatedSimple(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	listVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(0, 4)
	for range itemCount {
		if v := FieldValue(fd, opts.nested()); v != nil {
			listVal.List().Append(*v)
		}
	}
	return &listVal
}

// Repeated returns a fake repeated value given a field descriptor.
func Repeated(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	constraints, err := protovalidate.ResolveFieldRules(fd)
	if err != nil || constraints == nil {
		return repeatedSimple(msg, fd, opts)
	}
	rules := constraints.GetRepeated()
	if rules == nil {
		return repeatedSimple(msg, fd, opts)
	}
	min, max := uint64(0), uint64(4)
	if rules.MinItems != nil {
		min = rules.GetMinItems()
	}
	if rules.MaxItems != nil {
		max = rules.GetMaxItems()
	}

	// Ensure max is at least min, especially if min was set by rules and is > default max
	if min > max {
		max = min
	}

	listVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(int(min), int(max))

ItemLoop:
	for range itemCount {
		// Retry up to 20 times to generate a valid item.
		for range 20 { // Existing retry for valid item
			if v := FieldValue(fd, opts.nested().WithExtraFieldConstraints(rules.GetItems())); v != nil {
				if rules.GetUnique() { // Check for uniqueness rule
					isUnique := true
					for j := 0; j < listVal.List().Len(); j++ {
						// Compare based on the underlying value type
						existingVal := listVal.List().Get(j)
						if existingVal.IsValid() && v.IsValid() {
							switch fd.Kind() {
							case protoreflect.MessageKind:
								if proto.Equal(existingVal.Message().Interface(), v.Message().Interface()) {
									isUnique = false
								}
							case protoreflect.EnumKind, protoreflect.Int32Kind, protoreflect.Int64Kind,
								protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Sint32Kind,
								protoreflect.Sint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
								protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind, protoreflect.FloatKind,
								protoreflect.DoubleKind, protoreflect.BoolKind, protoreflect.StringKind,
								protoreflect.BytesKind:
								if existingVal.Interface() == v.Interface() { // Direct comparison for primitive types
									isUnique = false
								}
							}
						}
						if !isUnique {
							break
						}
					}
					if !isUnique {
						continue // Not unique, try generating another value
					}
				}
				listVal.List().Append(*v)
				continue ItemLoop // Success, move to the next item.
			}
		}
		// If we exhausted retries, we will fail to generate this item.
		// This may result in a list with fewer items than min_items if generation is impossible.
	}
	return &listVal
}
