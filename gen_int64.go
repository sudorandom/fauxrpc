package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type Int64Hints struct {
	Rules *validate.Int64Rules
}

func GenerateInt64(faker *gofakeit.Faker, hints Int64Hints) int64 {
	if hints.Rules == nil {
		return faker.Int64()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := int64(0), int64(math.MaxInt64)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.Int64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Int64Rules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.Int64Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.Int64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return int64(faker.IntRange(int(minVal), int(maxVal)))
}
