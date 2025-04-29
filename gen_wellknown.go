package fauxrpc

import (
	"strings"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"buf.build/go/protovalidate/resolve"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func durationSimple(opts GenOptions) *durationpb.Duration {
	duration := time.Duration(opts.fake().Uint64() % uint64(30*time.Hour*24))
	return durationpb.New(duration)
}

// GoogleDuration generates a random google.protobuf.Duration value.
func GoogleDuration(fd protoreflect.FieldDescriptor, opts GenOptions) *durationpb.Duration {
	constraints := resolve.FieldRules(fd)
	if constraints == nil {
		return durationSimple(opts)
	}
	rules := constraints.GetDuration()
	if rules == nil {
		return durationSimple(opts)
	}

	if rules.Const != nil {
		return rules.Const
	}
	if len(rules.Example) > 0 {
		return rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)]
	}

	minVal, maxVal := time.Duration(0), time.Duration(30*24*time.Hour)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.DurationRules_Gt:
			minVal = v.Gt.AsDuration() + 1
		case *validate.DurationRules_Gte:
			minVal = v.Gte.AsDuration()
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.DurationRules_Lt:
			maxVal = v.Lt.AsDuration() - 1
		case *validate.DurationRules_Lte:
			maxVal = v.Lte.AsDuration()
		}
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return durationpb.New(time.Duration(opts.fake().IntRange(int(minVal), int(maxVal))))
}

func generateTimestampSimple(opts GenOptions) *timestamppb.Timestamp {
	return timestamppb.New(opts.fake().Date())
}

// GoogleTimestamp generates a random google.protobuf.Timestamp value.
func GoogleTimestamp(fd protoreflect.FieldDescriptor, opts GenOptions) *timestamppb.Timestamp {
	constraints := resolve.FieldRules(fd)
	if constraints == nil {
		return generateTimestampSimple(opts)
	}
	rules := constraints.GetTimestamp()
	if rules == nil {
		return generateTimestampSimple(opts)
	}

	if rules.Const != nil {
		return rules.Const
	}
	if len(rules.Example) > 0 {
		return rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)]
	}

	minVal, maxVal := time.Now().Add(20*365*24*time.Hour), time.Now().Add(10*365*24*time.Hour)
	if rules.GreaterThan != nil {
		switch v := rules.GreaterThan.(type) {
		case *validate.TimestampRules_Gt:
			minVal = v.Gt.AsTime().Add(1)
		case *validate.TimestampRules_Gte:
			minVal = v.Gte.AsTime()
		}
	}
	if rules.LessThan != nil {
		switch v := rules.LessThan.(type) {
		case *validate.TimestampRules_Lt:
			maxVal = v.Lt.AsTime().Add(-1)
		case *validate.TimestampRules_Lte:
			maxVal = v.Lte.AsTime()
		}
	}
	if rules.Within != nil {
		minVal = time.Now().Add(-rules.Within.AsDuration())
		maxVal = time.Now().Add(rules.Within.AsDuration())
	}

	min := minVal.UnixNano()
	max := maxVal.UnixNano()

	delta := max - min

	return timestamppb.New(time.Unix(0, (opts.fake().Int64()%delta)+min))
}

func GoogleValue(fd protoreflect.FieldDescriptor, opts GenOptions) *structpb.Value {
	options := []func() *structpb.Value{
		func() *structpb.Value { return structpb.NewNullValue() },
		func() *structpb.Value { return structpb.NewBoolValue(Bool(fd, opts)) },
		func() *structpb.Value { return structpb.NewNumberValue(Float64(fd, opts)) },
		func() *structpb.Value { return structpb.NewStringValue(String(fd, opts)) },
		func() *structpb.Value {
			list := &structpb.ListValue{}
			itemCount := opts.fake().IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				list.Values = append(list.Values, GoogleValue(fd, opts.nested()))
			}
			return structpb.NewListValue(list)
		},
		func() *structpb.Value {
			obj := &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			}
			itemCount := opts.fake().IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				obj.Fields[strings.ToLower(opts.fake().Word())] = GoogleValue(fd, opts.nested())
			}
			return structpb.NewStructValue(obj)
		},
	}
	fn := options[opts.fake().IntRange(0, len(options)-1)]
	return fn()
}
