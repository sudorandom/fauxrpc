package protocel

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func newEnv(files *protoregistry.Files) (*cel.Env, error) {
	return cel.NewEnv(
		cel.TypeDescs(files),
		cel.Variable("req", cel.DynType),
		cel.Variable("field", cel.StringType),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType),
		cel.Function("gen_bool",
			cel.Overload("gen_bool_one_arg", []*cel.Type{cel.StringType}, types.BoolType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Bool(fauxrpc.Bool(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_bytes",
			cel.Overload("gen_bytes_one_arg", []*cel.Type{cel.StringType}, types.BytesType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Bytes(fauxrpc.Bytes(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_enum",
			cel.Overload("gen_enum_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.Enum(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_fixed32",
			cel.Overload("gen_fixed32_one_arg", []*cel.Type{cel.StringType}, types.UintType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Uint(fauxrpc.Fixed32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_fixed64",
			cel.Overload("gen_fixed64_one_arg", []*cel.Type{cel.StringType}, types.UintType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Uint(fauxrpc.Fixed64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_float32",
			cel.Overload("gen_float32_one_arg", []*cel.Type{cel.StringType}, types.DoubleType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Double(fauxrpc.Float32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_float64",
			cel.Overload("gen_float64_one_arg", []*cel.Type{cel.StringType}, types.DoubleType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Double(fauxrpc.Float64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_int32",
			cel.Overload("gen_int32_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.Int32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_int64",
			cel.Overload("gen_int64_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.Int64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sfixed32",
			cel.Overload("gen_sfixed32_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.SFixed32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sfixed64",
			cel.Overload("gen_sfixed64_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.SFixed64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sint32",
			cel.Overload("gen_sint32_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.SInt32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sint64",
			cel.Overload("gen_sint64_one_arg", []*cel.Type{cel.StringType}, types.IntType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Int(fauxrpc.SInt64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_string",
			cel.Overload("gen_string_one_arg", []*cel.Type{cel.StringType}, types.StringType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.String(fauxrpc.String(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_uint32",
			cel.Overload("gen_uint32_one_arg", []*cel.Type{cel.StringType}, types.UintType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Uint(fauxrpc.UInt32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_uint64",
			cel.Overload("gen_uint64_one_arg", []*cel.Type{cel.StringType}, types.UintType,
				cel.UnaryBinding(func(fieldName ref.Val) ref.Val {
					nameStr := fieldName.Value().(string)
					desc, err := files.FindDescriptorByName(protoreflect.FullName(nameStr))
					if err != nil {
						return types.NewErr(fmt.Sprintf("no descriptor found named '%s'", nameStr))
					}
					fd, ok := desc.(protoreflect.FieldDescriptor)
					if !ok {
						return types.NewErr(fmt.Sprintf("expected a field descriptor, got: '%T'", desc))
					}
					return types.Uint(fauxrpc.UInt64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
	)
}
