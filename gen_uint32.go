package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type UInt32Hints struct {
	Rules *validate.UInt32Rules
}

func GenerateUInt32(faker *gofakeit.Faker, hints UInt32Hints) uint32 {
	if hints.Rules == nil {
		return faker.Uint32()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
	}
	minVal, maxVal := uint32(0), uint32(math.MaxInt32)
	if hints.Rules.GreaterThan != nil {
		switch v := hints.Rules.GreaterThan.(type) {
		case *validate.UInt32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.UInt32Rules_Gte:
			minVal = v.Gte
		}
	}
	if hints.Rules.LessThan != nil {
		switch v := hints.Rules.LessThan.(type) {
		case *validate.UInt32Rules_Lt:
			maxVal = v.Lt + 1
		case *validate.UInt32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return uint32(faker.IntRange(int(minVal), int(maxVal)))
}
