package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type Float64Hints struct {
	Rules *validate.DoubleRules
}

func GenerateFloat64(faker *gofakeit.Faker, hints Float64Hints) float64 {
	if hints.Rules == nil {
		return faker.Float64()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := float64(0), float64(math.MaxFloat64)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.DoubleRules_Gt:
			minVal = v.Gt + 1
		case *validate.DoubleRules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.DoubleRules_Lt:
			maxVal = v.Lt + 1
		case *validate.DoubleRules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return faker.Float64Range(minVal, maxVal)
}
