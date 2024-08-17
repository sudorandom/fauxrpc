package fauxrpc

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateEnum returns a fake enum value given a field descriptor.
func GenerateEnum(fd protoreflect.FieldDescriptor) protoreflect.EnumNumber {
	values := fd.Enum().Values()
	idx := gofakeit.IntRange(0, values.Len()-1)
	return protoreflect.EnumNumber(idx)
}
