package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type Float32Hints struct {
	Rules *validate.FloatRules
}

func GenerateFloat32(faker *gofakeit.Faker, hints Float32Hints) float32 {
	if hints.Rules == nil {
		return faker.Float32()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := float32(0), float32(math.MaxFloat32)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.FloatRules_Gt:
			minVal = v.Gt + 1
		case *validate.FloatRules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.FloatRules_Lt:
			maxVal = v.Lt + 1
		case *validate.FloatRules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return faker.Float32Range(minVal, maxVal)
}
