package fauxrpc

import (
	"slices"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func enumSimple(fd protoreflect.FieldDescriptor, opts GenOptions) protoreflect.EnumNumber {
	values := fd.Enum().Values()
	idx := opts.fake().IntRange(0, values.Len()-1)
	return protoreflect.EnumNumber(values.Get(idx).Number())
}

// Enum returns a fake enum value given a field descriptor.
func Enum(fd protoreflect.FieldDescriptor, opts GenOptions) protoreflect.EnumNumber {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return enumSimple(fd, opts)
	}
	rules := constraints.GetEnum()
	if rules == nil {
		return enumSimple(fd, opts)
	}

	if rules.Const != nil {
		return protoreflect.EnumNumber(*rules.Const)
	}
	if len(rules.Example) > 0 {
		return protoreflect.EnumNumber(rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)])
	}

	// Collect all allowed enum values based on 'In' and 'NotIn' rules
	var allowedValues []protoreflect.EnumNumber
	allEnumValues := fd.Enum().Values()

	// If 'In' rule is present, only consider values in 'In' list
	if len(rules.In) > 0 {
		for _, val := range rules.In {
			// Check if this value is also forbidden by NotIn
			if !slices.Contains(rules.NotIn, val) {
				allowedValues = append(allowedValues, protoreflect.EnumNumber(val))
			}
		}
	} else {
		// If no 'In' rule, consider all enum values and filter out 'NotIn' values
		for i := 0; i < allEnumValues.Len(); i++ {
			val := allEnumValues.Get(i).Number()
			if !slices.Contains(rules.NotIn, int32(val)) {
				allowedValues = append(allowedValues, val)
			}
		}
	}

	if constraints.GetRequired() {
		allowedValues = slices.DeleteFunc(allowedValues, func(v protoreflect.EnumNumber) bool {
			return v == 0
		})
	}

	if len(allowedValues) > 0 {
		// Pick a random value from the allowed list
		return allowedValues[opts.fake().IntRange(0, len(allowedValues)-1)]
	} else {
		// This is an impossible scenario: no valid enum value can be generated.
		// This indicates a misconfigured validation rule in the proto definition.
		// For now, we will return the first enum value, which might still be invalid
		// according to the rules, but prevents a crash.
		// A proper fix would involve reporting this as an error upstream or
		// having the validation library prevent such impossible rules.
		if allEnumValues.Len() > 0 {
			return allEnumValues.Get(0).Number()
		}
		return 0 // Default to 0 if no enum values exist at all
	}
}
