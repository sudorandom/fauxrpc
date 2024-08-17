package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GenerateSInt64(fd protoreflect.FieldDescriptor) int64 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Int64()
	}
	rules := constraints.GetSint64()
	if rules == nil {
		return gofakeit.Int64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int64(0), int64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SInt64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SInt64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SInt64Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.SInt64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return int64(gofakeit.IntRange(int(minVal), int(maxVal)))
}
