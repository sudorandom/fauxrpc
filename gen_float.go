package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Float32 returns a fake float32 value given a field descriptor.
func Float32(fd protoreflect.FieldDescriptor, opts GenOptions) float32 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return gofakeit.Float32()
	}
	rules := constraints.GetFloat()
	if rules == nil {
		return gofakeit.Float32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := float32(0), float32(math.MaxFloat32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.FloatRules_Gt:
			minVal = v.Gt + 1
		case *validate.FloatRules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.FloatRules_Lt:
			maxVal = v.Lt - 1
		case *validate.FloatRules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return gofakeit.Float32Range(minVal, maxVal)
}

// Float64 returns a fake float64 value given a field descriptor.
func Float64(fd protoreflect.FieldDescriptor, opts GenOptions) float64 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return gofakeit.Float64()
	}
	rules := constraints.GetDouble()
	if rules == nil {
		return gofakeit.Float64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := float64(0), float64(math.MaxFloat64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.DoubleRules_Gt:
			minVal = v.Gt + 1
		case *validate.DoubleRules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.DoubleRules_Lt:
			maxVal = v.Lt - 1
		case *validate.DoubleRules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return gofakeit.Float64Range(minVal, maxVal)
}
