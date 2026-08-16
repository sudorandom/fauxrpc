package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	minInt32 = math.MinInt32
	maxInt32 = math.MaxInt32
	// maxViolatingLength bounds how long a generated string or byte slice may
	// get while violating a length rule.
	maxViolatingLength = 4096
)

type numeric interface {
	~int32 | ~int64 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// numericRules is the subset of a protovalidate numeric rule set that a single
// value can violate, normalized across the twelve numeric rule messages.
type numericRules[T numeric] struct {
	constVal *T
	in       []T
	notIn    []T
	gt       *T
	gte      *T
	lt       *T
	lte      *T
}

// violateNumeric rolls each rule against ViolateRules and returns a value that
// fails one of the rules that came up.
func violateNumeric[T numeric](rules numericRules[T], lowest, highest T, opts GenOptions) (T, bool) {
	return numericViolations(rules, lowest, highest).pick(opts)
}

// numericViolations groups counterexamples by the rule each one breaks. lowest
// and highest bound T so that stepping outside a range cannot wrap around.
func numericViolations[T numeric](rules numericRules[T], lowest, highest T) ruleViolations[T] {
	var violations ruleViolations[T]
	if rules.constVal != nil {
		var anythingElse []T
		if *rules.constVal < highest {
			anythingElse = append(anythingElse, *rules.constVal+1)
		}
		if *rules.constVal > lowest {
			anythingElse = append(anythingElse, *rules.constVal-1)
		}
		violations = violations.add(anythingElse...)
	}
	violations = violations.add(rules.notIn...)
	if len(rules.in) > 0 {
		if v, ok := valueNotIn(rules.in, lowest, highest); ok {
			violations = violations.add(v)
		}
	}
	// A bound is not greater/less than itself, so it is its own counterexample
	// for the exclusive forms.
	if rules.gt != nil {
		violations = violations.add(*rules.gt)
	}
	if rules.lt != nil {
		violations = violations.add(*rules.lt)
	}
	if rules.gte != nil && *rules.gte > lowest {
		violations = violations.add(*rules.gte - 1)
	}
	if rules.lte != nil && *rules.lte < highest {
		violations = violations.add(*rules.lte + 1)
	}
	return violations
}

// valueNotIn returns a value missing from in. Scanning len(in)+1 candidates
// always finds one unless they fall outside the type's range.
func valueNotIn[T numeric](in []T, lowest, highest T) (T, bool) {
	allowed := make(map[T]struct{}, len(in))
	for _, v := range in {
		allowed[v] = struct{}{}
	}
	for i := range len(in) + 1 {
		candidate := T(i)
		if candidate > highest || candidate < lowest {
			break
		}
		if _, ok := allowed[candidate]; !ok {
			return candidate, true
		}
	}
	return 0, false
}

func violatingNumericValue(fd protoreflect.FieldDescriptor, rules *validate.FieldRules, opts GenOptions) (protoreflect.Value, bool) {
	switch fd.Kind() {
	case protoreflect.Int32Kind:
		if v, ok := violateInt32Rules(rules.GetInt32(), opts); ok {
			return protoreflect.ValueOfInt32(v), true
		}
	case protoreflect.Sint32Kind:
		if v, ok := violateSInt32Rules(rules.GetSint32(), opts); ok {
			return protoreflect.ValueOfInt32(v), true
		}
	case protoreflect.Sfixed32Kind:
		if v, ok := violateSFixed32Rules(rules.GetSfixed32(), opts); ok {
			return protoreflect.ValueOfInt32(v), true
		}
	case protoreflect.Int64Kind:
		if v, ok := violateInt64Rules(rules.GetInt64(), opts); ok {
			return protoreflect.ValueOfInt64(v), true
		}
	case protoreflect.Sint64Kind:
		if v, ok := violateSInt64Rules(rules.GetSint64(), opts); ok {
			return protoreflect.ValueOfInt64(v), true
		}
	case protoreflect.Sfixed64Kind:
		if v, ok := violateSFixed64Rules(rules.GetSfixed64(), opts); ok {
			return protoreflect.ValueOfInt64(v), true
		}
	case protoreflect.Uint32Kind:
		if v, ok := violateUInt32Rules(rules.GetUint32(), opts); ok {
			return protoreflect.ValueOfUint32(v), true
		}
	case protoreflect.Fixed32Kind:
		if v, ok := violateFixed32Rules(rules.GetFixed32(), opts); ok {
			return protoreflect.ValueOfUint32(v), true
		}
	case protoreflect.Uint64Kind:
		if v, ok := violateUInt64Rules(rules.GetUint64(), opts); ok {
			return protoreflect.ValueOfUint64(v), true
		}
	case protoreflect.Fixed64Kind:
		if v, ok := violateFixed64Rules(rules.GetFixed64(), opts); ok {
			return protoreflect.ValueOfUint64(v), true
		}
	case protoreflect.FloatKind:
		if v, ok := violateFloatRules(rules.GetFloat(), opts); ok {
			return protoreflect.ValueOfFloat32(v), true
		}
	case protoreflect.DoubleKind:
		if v, ok := violateDoubleRules(rules.GetDouble(), opts); ok {
			return protoreflect.ValueOfFloat64(v), true
		}
	}
	return protoreflect.Value{}, false
}

func isZeroNumber(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return val.Int() == 0
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return val.Uint() == 0
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return val.Float() == 0
	}
	return false
}

func violateInt32Rules(rules *validate.Int32Rules, opts GenOptions) (int32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.Int32Rules_Gt:
		n.gt = &v.Gt
	case *validate.Int32Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.Int32Rules_Lt:
		n.lt = &v.Lt
	case *validate.Int32Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt32, math.MaxInt32, opts)
}

func violateSInt32Rules(rules *validate.SInt32Rules, opts GenOptions) (int32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.SInt32Rules_Gt:
		n.gt = &v.Gt
	case *validate.SInt32Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.SInt32Rules_Lt:
		n.lt = &v.Lt
	case *validate.SInt32Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt32, math.MaxInt32, opts)
}

func violateSFixed32Rules(rules *validate.SFixed32Rules, opts GenOptions) (int32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.SFixed32Rules_Gt:
		n.gt = &v.Gt
	case *validate.SFixed32Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.SFixed32Rules_Lt:
		n.lt = &v.Lt
	case *validate.SFixed32Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt32, math.MaxInt32, opts)
}

func violateInt64Rules(rules *validate.Int64Rules, opts GenOptions) (int64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.Int64Rules_Gt:
		n.gt = &v.Gt
	case *validate.Int64Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.Int64Rules_Lt:
		n.lt = &v.Lt
	case *validate.Int64Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt64, math.MaxInt64, opts)
}

func violateSInt64Rules(rules *validate.SInt64Rules, opts GenOptions) (int64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.SInt64Rules_Gt:
		n.gt = &v.Gt
	case *validate.SInt64Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.SInt64Rules_Lt:
		n.lt = &v.Lt
	case *validate.SInt64Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt64, math.MaxInt64, opts)
}

func violateSFixed64Rules(rules *validate.SFixed64Rules, opts GenOptions) (int64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[int64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.SFixed64Rules_Gt:
		n.gt = &v.Gt
	case *validate.SFixed64Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.SFixed64Rules_Lt:
		n.lt = &v.Lt
	case *validate.SFixed64Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, math.MinInt64, math.MaxInt64, opts)
}

func violateUInt32Rules(rules *validate.UInt32Rules, opts GenOptions) (uint32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[uint32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.UInt32Rules_Gt:
		n.gt = &v.Gt
	case *validate.UInt32Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.UInt32Rules_Lt:
		n.lt = &v.Lt
	case *validate.UInt32Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, 0, math.MaxUint32, opts)
}

func violateFixed32Rules(rules *validate.Fixed32Rules, opts GenOptions) (uint32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[uint32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.Fixed32Rules_Gt:
		n.gt = &v.Gt
	case *validate.Fixed32Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.Fixed32Rules_Lt:
		n.lt = &v.Lt
	case *validate.Fixed32Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, 0, math.MaxUint32, opts)
}

func violateUInt64Rules(rules *validate.UInt64Rules, opts GenOptions) (uint64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[uint64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.UInt64Rules_Gt:
		n.gt = &v.Gt
	case *validate.UInt64Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.UInt64Rules_Lt:
		n.lt = &v.Lt
	case *validate.UInt64Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, 0, math.MaxUint64, opts)
}

func violateFixed64Rules(rules *validate.Fixed64Rules, opts GenOptions) (uint64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[uint64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.Fixed64Rules_Gt:
		n.gt = &v.Gt
	case *validate.Fixed64Rules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.Fixed64Rules_Lt:
		n.lt = &v.Lt
	case *validate.Fixed64Rules_Lte:
		n.lte = &v.Lte
	}
	return violateNumeric(n, 0, math.MaxUint64, opts)
}

func violateFloatRules(rules *validate.FloatRules, opts GenOptions) (float32, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[float32]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.FloatRules_Gt:
		n.gt = &v.Gt
	case *validate.FloatRules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.FloatRules_Lt:
		n.lt = &v.Lt
	case *validate.FloatRules_Lte:
		n.lte = &v.Lte
	}
	violations := numericViolations(n, -math.MaxFloat32, math.MaxFloat32)
	if rules.GetFinite() {
		violations = violations.add(
			float32(math.Inf(1)),
			float32(math.Inf(-1)),
			float32(math.NaN()),
		)
	}
	return violations.pick(opts)
}

func violateDoubleRules(rules *validate.DoubleRules, opts GenOptions) (float64, bool) {
	if rules == nil {
		return 0, false
	}
	n := numericRules[float64]{constVal: rules.Const, in: rules.In, notIn: rules.NotIn}
	switch v := rules.GreaterThan.(type) {
	case *validate.DoubleRules_Gt:
		n.gt = &v.Gt
	case *validate.DoubleRules_Gte:
		n.gte = &v.Gte
	}
	switch v := rules.LessThan.(type) {
	case *validate.DoubleRules_Lt:
		n.lt = &v.Lt
	case *validate.DoubleRules_Lte:
		n.lte = &v.Lte
	}
	violations := numericViolations(n, -math.MaxFloat64, math.MaxFloat64)
	if rules.GetFinite() {
		violations = violations.add(math.Inf(1), math.Inf(-1), math.NaN())
	}
	return violations.pick(opts)
}
