package fauxrpc

import (
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func mapSimple(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	mapVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(0, 4)
	for i := 0; i < itemCount; i++ {
		v := FieldValue(fd.MapKey(), opts)
		w := FieldValue(fd.MapValue(), opts)
		if v != nil && w != nil {
			mapVal.Map().Set((*v).MapKey(), *w)
		}
	}
	return &mapVal
}

// Map returns a fake repeated value given a field descriptor.
func Map(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	if opts.MaxDepth <= 0 {
		return nil
	}
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return mapSimple(msg, fd, opts)
	}
	rules := constraints.GetMap()
	if rules == nil {
		return mapSimple(msg, fd, opts)
	}
	min, max := uint64(0), uint64(4)
	if rules.MinPairs != nil {
		min = rules.GetMinPairs()
	}
	if rules.MaxPairs != nil {
		max = rules.GetMaxPairs()
	}

	// Ensure max is at least min
	if min > max {
		max = min
	}

	mapVal := msg.NewField(fd)
	itemCount := opts.fake().IntRange(int(min), int(max))
	for i := 0; i < itemCount; i++ {
		v := FieldValue(fd.MapKey(), opts.WithExtraFieldConstraints(rules.Keys))
		w := FieldValue(fd.MapValue(), opts.WithExtraFieldConstraints(rules.Values))
		if v != nil && w != nil {
			mapVal.Map().Set((*v).MapKey(), *w)
		} else {
			slog.Warn(fmt.Sprintf("Unknown map k/v %s %v", fd.FullName(), fd.Kind()))
		}
	}
	return &mapVal
}
