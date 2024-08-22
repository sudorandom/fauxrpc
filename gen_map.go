package fauxrpc

import (
	"fmt"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func mapSimple(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	mapVal := msg.NewField(fd)
	itemCount := gofakeit.IntRange(0, 4)
	for i := 0; i < itemCount; i++ {
		v := getFieldValue(fd.MapKey(), opts.nested())
		w := getFieldValue(fd.MapValue(), opts.nested())
		if v != nil && w != nil {
			mapVal.Map().Set((*v).MapKey(), *w)
		} else {
			slog.Warn(fmt.Sprintf("Unknown map k/v %s %v", fd.FullName(), fd.Kind()))
		}
	}
	return &mapVal
}

// Map returns a fake repeated value given a field descriptor.
func Map(msg protoreflect.Message, fd protoreflect.FieldDescriptor, opts GenOptions) *protoreflect.Value {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return mapSimple(msg, fd, opts)
	}
	rules := constraints.GetEnum()
	if rules == nil {
		return mapSimple(msg, fd, opts)
	}
	min, max := uint64(0), uint64(4)
	if constraints.GetMap().MinPairs != nil {
		min = constraints.GetMap().GetMinPairs()
	}
	if constraints.GetMap().MaxPairs != nil {
		max = constraints.GetMap().GetMaxPairs()
	}

	mapVal := msg.NewField(fd)
	itemCount := gofakeit.IntRange(int(min), int(max))
	for i := 0; i < itemCount; i++ {
		v := getFieldValue(fd.MapKey(), opts.nested().withExtraFieldConstraints(constraints.GetMap().Keys))
		w := getFieldValue(fd.MapValue(), opts.nested().withExtraFieldConstraints(constraints.GetMap().Values))
		if v != nil && w != nil {
			mapVal.Map().Set((*v).MapKey(), *w)
		} else {
			slog.Warn(fmt.Sprintf("Unknown map k/v %s %v", fd.FullName(), fd.Kind()))
		}
	}
	return &mapVal
}
