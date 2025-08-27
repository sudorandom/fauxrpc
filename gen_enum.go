package fauxrpc

import (
	"slices"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func Enum(fd protoreflect.FieldDescriptor, opts GenOptions) protoreflect.EnumNumber {
	constraints := getFieldConstraints(fd, opts)
	allEnumValues := fd.Enum().Values()
	allowedValues := []protoreflect.EnumNumber{}

	if allEnumValues.Len() > 0 {
		for i := 0; i < allEnumValues.Len(); i++ {
			allowedValues = append(allowedValues, allEnumValues.Get(i).Number())
		}
	}

	if constraints != nil {
		if enumRules := constraints.GetEnum(); enumRules != nil {
			if enumRules.Const != nil {
				return protoreflect.EnumNumber(*enumRules.Const)
			}
			if len(enumRules.In) > 0 {
				inValues := make(map[protoreflect.EnumNumber]struct{})
				for _, v := range enumRules.In {
					inValues[protoreflect.EnumNumber(v)] = struct{}{}
				}
				allowedValues = slices.DeleteFunc(allowedValues, func(v protoreflect.EnumNumber) bool {
					_, ok := inValues[v]
					return !ok
				})
			}
			if len(enumRules.NotIn) > 0 {
				notInValues := make(map[protoreflect.EnumNumber]struct{})
				for _, v := range enumRules.NotIn {
					notInValues[protoreflect.EnumNumber(v)] = struct{}{}
				}
				allowedValues = slices.DeleteFunc(allowedValues, func(v protoreflect.EnumNumber) bool {
					_, ok := notInValues[v]
					return ok
				})
			}
		}
		if constraints.GetRequired() {
			allowedValues = slices.DeleteFunc(allowedValues, func(v protoreflect.EnumNumber) bool {
				return v == 0
			})
		}
	}

	if len(allowedValues) > 0 {
		return allowedValues[opts.fake().IntRange(0, len(allowedValues)-1)]
	}

	if allEnumValues.Len() > 0 {
		idx := opts.fake().IntRange(0, allEnumValues.Len()-1)
		return allEnumValues.Get(idx).Number()
	}
	return 0
}
