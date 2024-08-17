package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateUInt64 returns a fake uint64 value given a field descriptor.
func GenerateUInt64(fd protoreflect.FieldDescriptor) uint64 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint64()
	}
	rules := constraints.GetUint64()
	if rules == nil {
		return gofakeit.Uint64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint64(0), uint64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.UInt64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.UInt64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.UInt64Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.UInt64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint64(gofakeit.UintRange(uint(minVal), uint(maxVal)))
}
