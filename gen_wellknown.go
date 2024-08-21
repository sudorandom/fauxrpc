package fauxrpc

import (
	"strings"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func durationSimple() *durationpb.Duration {
	duration := time.Duration(gofakeit.Uint64() % uint64(30*time.Hour*24))
	return durationpb.New(duration)
}

// GoogleDuration generates a random google.protobuf.Duration value.
func GoogleDuration(fd protoreflect.FieldDescriptor) *durationpb.Duration {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return durationSimple()
	}
	rules := constraints.GetDuration()
	if rules == nil {
		return durationSimple()
	}

	if rules.Const != nil {
		return rules.Const
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
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return durationpb.New(time.Duration(gofakeit.IntRange(int(minVal), int(maxVal))))
}

func generateTimestampSimple() *timestamppb.Timestamp {
	return timestamppb.New(gofakeit.Date())
}

// GoogleTimestamp generates a random google.protobuf.Timestamp value.
func GoogleTimestamp(fd protoreflect.FieldDescriptor) *timestamppb.Timestamp {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return generateTimestampSimple()
	}
	rules := constraints.GetTimestamp()
	if rules == nil {
		return generateTimestampSimple()
	}

	if rules.Const != nil {
		return rules.Const
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

	return timestamppb.New(time.Unix(0, (gofakeit.Int64()%delta)+min))
}

func GoogleValue(fd protoreflect.FieldDescriptor, opts GenOptions) *structpb.Value {
	options := []func() *structpb.Value{
		func() *structpb.Value { return structpb.NewNullValue() },
		func() *structpb.Value { return structpb.NewBoolValue(Bool(fd)) },
		func() *structpb.Value { return structpb.NewNumberValue(Float64(fd)) },
		func() *structpb.Value { return structpb.NewStringValue(String(fd)) },
		func() *structpb.Value {
			list := &structpb.ListValue{}
			itemCount := gofakeit.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				list.Values = append(list.Values, GoogleValue(fd, opts.nested()))
			}
			return structpb.NewListValue(list)
		},
		func() *structpb.Value {
			obj := &structpb.Struct{}
			itemCount := gofakeit.IntRange(0, 4)
			for i := 0; i < itemCount; i++ {
				obj.Fields[strings.ToLower(gofakeit.Word())] = GoogleValue(fd, opts.nested())
			}
			return structpb.NewStructValue(obj)
		},
	}
	fn := options[gofakeit.IntRange(0, len(options)-1)]
	return fn()
}
