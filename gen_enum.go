package fauxrpc

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func enumSimple(fd protoreflect.FieldDescriptor) protoreflect.EnumNumber {
	values := fd.Enum().Values()
	idx := gofakeit.IntRange(0, values.Len()-1)
	return protoreflect.EnumNumber(idx)
}

// Enum returns a fake enum value given a field descriptor.
func Enum(fd protoreflect.FieldDescriptor) protoreflect.EnumNumber {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return enumSimple(fd)
	}
	rules := constraints.GetEnum()
	if rules == nil {
		return enumSimple(fd)
	}

	if rules.Const != nil {
		return protoreflect.EnumNumber(*rules.Const)
	}

	if len(rules.In) > 0 {
		return protoreflect.EnumNumber(rules.In[gofakeit.IntRange(0, len(rules.In)-1)])
	}

	return enumSimple(fd)
}
