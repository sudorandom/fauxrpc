package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

func enumSimple(fd protoreflect.FieldDescriptor, opts GenOptions) protoreflect.EnumNumber {
	values := fd.Enum().Values()
	idx := opts.fake().IntRange(0, values.Len()-1)
	return protoreflect.EnumNumber(idx)
}

// Enum returns a fake enum value given a field descriptor.
func Enum(fd protoreflect.FieldDescriptor, opts GenOptions) protoreflect.EnumNumber {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return enumSimple(fd, opts)
	}
	rules := constraints.GetEnum()
	if rules == nil {
		return enumSimple(fd, opts)
	}

	if rules.Const != nil {
		return protoreflect.EnumNumber(*rules.Const)
	}
	if len(rules.Example) > 0 {
		return protoreflect.EnumNumber(rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)])
	}

	if len(rules.In) > 0 {
		return protoreflect.EnumNumber(rules.In[opts.fake().IntRange(0, len(rules.In)-1)])
	}

	return enumSimple(fd, opts)
}
