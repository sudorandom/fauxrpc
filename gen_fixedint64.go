package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type Fixed64Hints struct {
	Rules *validate.Fixed64Rules
}

func GenerateFixed64(faker *gofakeit.Faker, hints Fixed64Hints) uint64 {
	if hints.Rules == nil {
		return faker.Uint64()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := uint64(0), uint64(math.MaxInt64)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.Fixed64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Fixed64Rules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.Fixed64Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.Fixed64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return uint64(faker.UintRange(uint(minVal), uint(maxVal)))
}
