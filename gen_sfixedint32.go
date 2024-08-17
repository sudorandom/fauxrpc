package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type SFixedInt32Hints struct {
	Rules *validate.SFixed32Rules
}

func GenerateSFixedInt32(faker *gofakeit.Faker, hints SFixedInt32Hints) int32 {
	if hints.Rules == nil {
		return faker.Int32()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := int32(0), int32(math.MaxInt32)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.SFixed32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SFixed32Rules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.SFixed32Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.SFixed32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return int32(faker.IntRange(int(minVal), int(maxVal)))
}