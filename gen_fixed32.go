package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateFixed32 returns a fake fixed32 value given a field descriptor.
func GenerateFixed32(fd protoreflect.FieldDescriptor) uint32 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint32()
	}
	rules := constraints.GetFixed32()
	if rules == nil {
		return gofakeit.Uint32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint32(0), uint32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.Fixed32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Fixed32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.Fixed32Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.Fixed32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint32(gofakeit.IntRange(int(minVal), int(maxVal)))
}
