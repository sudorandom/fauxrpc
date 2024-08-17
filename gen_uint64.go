package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type UInt64Hints struct {
	Rules *validate.UInt64Rules
}

func GenerateUInt64(faker *gofakeit.Faker, hints UInt64Hints) uint64 {
	if hints.Rules == nil {
		return faker.Uint64()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := uint64(0), uint64(math.MaxInt64)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.UInt64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.UInt64Rules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.UInt64Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.UInt64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return uint64(faker.UintRange(uint(minVal), uint(maxVal)))
}
