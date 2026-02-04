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

	seen := make(map[any]struct{})
ItemLoop:
	for range itemCount {
		// Retry up to 20 times to generate a valid item.
		for range 20 { // Existing retry for valid item
			if v := FieldValue(fd, opts.nested().WithExtraFieldConstraints(rules.GetItems())); v != nil {
				if rules.GetUnique() { // Check for uniqueness rule
					var key any
					switch fd.Kind() {
					case protoreflect.MessageKind, protoreflect.GroupKind:
						// Marshal to bytes for uniqueness check for messages.
						// We use deterministic marshaling to ensure consistent bytes for equal messages.
						b, err := proto.MarshalOptions{Deterministic: true}.Marshal(v.Message().Interface())
						if err == nil {
							key = string(b)
						}
					case protoreflect.BytesKind:
						key = string(v.Bytes())
					default:
						key = v.Interface()
					}

					if key != nil {
						if _, exists := seen[key]; exists {
							continue // Not unique, try generating another value
						}
						seen[key] = struct{}{}
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
