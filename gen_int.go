package fauxrpc

import (
	"math"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Int32 returns a fake int32 value given a field descriptor.
func Int32(fd protoreflect.FieldDescriptor, opts GenOptions) int32 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int32()
	}
	rules := constraints.GetInt32()
	if rules == nil {
		return opts.fake().Int32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	if len(rules.Example) > 0 {
		return rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)]
	}

	minVal, maxVal := int32(0), int32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.Int32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Int32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.Int32Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.Int32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int32(opts.fake().IntRange(int(minVal), int(maxVal)))
}

// Int64 returns a fake int64 value given a field descriptor.
func Int64(fd protoreflect.FieldDescriptor, opts GenOptions) int64 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int64()
	}
	rules := constraints.GetInt64()
	if rules == nil {
		return opts.fake().Int64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int64(0), int64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.Int64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.Int64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.Int64Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.Int64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int64(opts.fake().IntRange(int(minVal), int(maxVal)))
}

// SInt32 returns a fake sint32 value given a field descriptor.
func SInt32(fd protoreflect.FieldDescriptor, opts GenOptions) int32 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int32()
	}
	rules := constraints.GetSint32()
	if rules == nil {
		return opts.fake().Int32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int32(0), int32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SInt32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SInt32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SInt32Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.SInt32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int32(opts.fake().IntRange(int(minVal), int(maxVal)))
}

// SInt64 returns a fake sint64 value given a field descriptor.
func SInt64(fd protoreflect.FieldDescriptor, opts GenOptions) int64 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int64()
	}
	rules := constraints.GetSint64()
	if rules == nil {
		return opts.fake().Int64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int64(0), int64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SInt64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SInt64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SInt64Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.SInt64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int64(opts.fake().IntRange(int(minVal), int(maxVal)))
}

// SFixed32 returns a fake sfixedint32 value given a field descriptor.
func SFixed32(fd protoreflect.FieldDescriptor, opts GenOptions) int32 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int32()
	}
	rules := constraints.GetSfixed32()
	if rules == nil {
		return opts.fake().Int32()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int32(0), int32(math.MaxInt32)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SFixed32Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SFixed32Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SFixed32Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.SFixed32Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int32(opts.fake().IntRange(int(minVal), int(maxVal)))
}

// SFixed64 returns a fake sfixed64 value given a field descriptor.
func SFixed64(fd protoreflect.FieldDescriptor, opts GenOptions) int64 {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return opts.fake().Int64()
	}
	rules := constraints.GetSfixed64()
	if rules == nil {
		return opts.fake().Int64()
	}

	if rules.Const != nil {
		return *rules.Const
	}
	minVal, maxVal := int64(0), int64(math.MaxInt64)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.SFixed64Rules_Gt:
			minVal = v.Gt + 1
		case *validate.SFixed64Rules_Gte:
			minVal = v.Gte
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.SFixed64Rules_Lt:
			maxVal = v.Lt - 1
		case *validate.SFixed64Rules_Lte:
			maxVal = v.Lte
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return int64(opts.fake().IntRange(int(minVal), int(maxVal)))
}
