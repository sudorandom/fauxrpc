package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateSFixed32 returns a fake sfixedint32 value given a field descriptor.
func GenerateSFixed32(fd protoreflect.FieldDescriptor) int32 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Int32()
	}
	rules := constraints.GetSfixed32()
	if rules == nil {
		return gofakeit.Int32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int32(0), int32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SFixed32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SFixed32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SFixed32Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.SFixed32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return int32(gofakeit.IntRange(int(minVal), int(maxVal)))
}
