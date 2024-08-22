package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Bool returns a fake boolean value given a field descriptor.
func Bool(fd protoreflect.FieldDescriptor, opts GenOptions) bool {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return true
	}
	rules := constraints.GetBool()
	if rules == nil {
		return true
	}
	if rules.Const != nil {
		return *rules.Const
	}
	return true
}
