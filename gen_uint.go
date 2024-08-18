package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// UInt32 returns a fake uint32 value given a field descriptor.
func UInt32(fd protoreflect.FieldDescriptor) uint32 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint32()
	}
	rules := constraints.GetUint32()
	if rules == nil {
		return gofakeit.Uint32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint32(0), uint32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.UInt32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.UInt32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.UInt32Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.UInt32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint32(gofakeit.IntRange(int(minVal), int(maxVal)))
}

// UInt64 returns a fake uint64 value given a field descriptor.
func UInt64(fd protoreflect.FieldDescriptor) uint64 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint64()
	}
	rules := constraints.GetUint64()
	if rules == nil {
		return gofakeit.Uint64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint64(0), uint64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.UInt64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.UInt64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.UInt64Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.UInt64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint64(gofakeit.UintRange(uint(minVal), uint(maxVal)))
}

// Fixed32 returns a fake fixed32 value given a field descriptor.
func Fixed32(fd protoreflect.FieldDescriptor) uint32 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint32()
	}
	rules := constraints.GetFixed32()
	if rules == nil {
		return gofakeit.Uint32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint32(0), uint32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.Fixed32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Fixed32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.Fixed32Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.Fixed32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint32(gofakeit.IntRange(int(minVal), int(maxVal)))
}

// Fixed64 returns a fake fixed64 value given a field descriptor.
func Fixed64(fd protoreflect.FieldDescriptor) uint64 {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return gofakeit.Uint64()
	}
	rules := constraints.GetFixed64()
	if rules == nil {
		return gofakeit.Uint64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := uint64(0), uint64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.Fixed64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Fixed64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.Fixed64Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.Fixed64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return uint64(gofakeit.UintRange(uint(minVal), uint(maxVal)))
}
