package fauxrpc

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DynamicStruct interface {
	NewMessage(opts GenOptions) (proto.Message, error)
	SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error
}

var _ DynamicStruct = (*dynamicStruct)(nil)

type dynamicStruct struct {
	messageDescriptor protoreflect.MessageDescriptor
	fields            map[protoreflect.FieldDescriptor]cel.Program
}

func NewDynamicStruct(md protoreflect.MessageDescriptor, fields map[string]string) (*dynamicStruct, error) {
	celFields := map[protoreflect.FieldDescriptor]cel.Program{}
	protoFields := md.Fields()
	for key, expr := range fields {
		field := getFieldFromName(protoFields, key)
		if field == nil {
			return nil, fmt.Errorf("field %s not found on %s", key, md.FullName())
		}

		env, err := cel.NewEnv(
			cel.Variable("req", cel.ObjectType(string(md.FullName()))),
			cel.Variable("service", cel.StringType),
			cel.Variable("method", cel.StringType),
			cel.Variable("procedure", cel.StringType),
			cel.Function("gen_bool",
				cel.Overload("gen_bool_noargs", []*cel.Type{}, types.BoolType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Bool(Bool(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_bytes",
				cel.Overload("gen_bytes_noargs", []*cel.Type{}, types.BytesType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Bytes(Bytes(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_enum",
				cel.Overload("gen_enum_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(Enum(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_fixed32",
				cel.Overload("gen_fixed32_noargs", []*cel.Type{}, types.UintType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Uint(Fixed32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_fixed64",
				cel.Overload("gen_fixed64_noargs", []*cel.Type{}, types.UintType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Uint(Fixed64(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_float32",
				cel.Overload("gen_float32_noargs", []*cel.Type{}, types.DoubleType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Double(Float32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_float64",
				cel.Overload("gen_float64_noargs", []*cel.Type{}, types.DoubleType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Double(Float64(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_int32",
				cel.Overload("gen_int32_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(Int32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_int64",
				cel.Overload("gen_int64_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(Int64(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_sfixed32",
				cel.Overload("gen_sfixed32_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(SFixed32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_sfixed64",
				cel.Overload("gen_sfixed64_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(SFixed64(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_sint32",
				cel.Overload("gen_sint32_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(SInt32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_sint64",
				cel.Overload("gen_sint64_noargs", []*cel.Type{}, types.IntType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Int(SInt64(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_string",
				cel.Overload("gen_string_noargs", []*cel.Type{}, types.StringType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.String(String(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_uint32",
				cel.Overload("gen_uint32_noargs", []*cel.Type{}, types.UintType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Uint(UInt32(field, GenOptions{}))
					}),
				),
			),
			cel.Function("gen_uint64",
				cel.Overload("gen_uint64_noargs", []*cel.Type{}, types.UintType,
					cel.FunctionBinding(func(values ...ref.Val) ref.Val {
						return types.Uint(UInt64(field, GenOptions{}))
					}),
				),
			),
		)
		if err != nil {
			return nil, err
		}

		ast, issues := env.Compile(expr)
		if issues != nil {
			return nil, issues.Err()
		}
		if err := checkCelType(field, ast.OutputType()); err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		program, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		celFields[field] = program
	}

	return &dynamicStruct{
		messageDescriptor: md,
		fields:            celFields,
	}, nil
}

// NewMessage implements DynamicStruct.
func (d *dynamicStruct) NewMessage(opts GenOptions) (protoreflect.ProtoMessage, error) {
	fmt.Println("NewMessage")
	msg := newMessage(d.messageDescriptor).Interface()
	if err := d.SetDataOnMessage(msg, opts); err != nil {
		return nil, err
	}
	return msg, nil
}

// SetDataOnMessage implements DynamicStruct.
func (d *dynamicStruct) SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error {
	for field, program := range d.fields {
		val, _, err := program.Eval(map[string]any{})
		if err != nil {
			return err
		}

		msg.ProtoReflect().Set(field, protoreflect.ValueOf(val.Value()))
	}
	return nil
}

func fieldToCELTypes(md protoreflect.FieldDescriptor) []*types.Type {
	switch md.Kind() {
	case protoreflect.BoolKind:
		return []*types.Type{types.BoolType}
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind,
		protoreflect.Sint64Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return []*types.Type{types.IntType}
	case protoreflect.EnumKind:
		return []*types.Type{types.IntType, types.StringType, types.UintType}
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		return []*types.Type{types.UintType}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return []*types.Type{types.DoubleType}
	case protoreflect.StringKind:
		return []*types.Type{types.StringType}
	case protoreflect.BytesKind:
		return []*types.Type{types.BytesType}
	default:
		return nil
	}
}

func checkCelType(md protoreflect.FieldDescriptor, t *types.Type) error {
	validTypes := fieldToCELTypes(md)
	if validTypes == nil {
		return fmt.Errorf("unhandled kind: %v", md.Kind())
	}

	for _, validType := range validTypes {
		if t == validType {
			return nil
		}
	}
	return fmt.Errorf("expected %v; got %s", validTypes, t)
}

func getFieldFromName(fds protoreflect.FieldDescriptors, key string) protoreflect.FieldDescriptor {
	if field := fds.ByName(protoreflect.Name(key)); field != nil {
		return field
	}
	if field := fds.ByTextName(key); field != nil {
		return field
	}
	if field := fds.ByJSONName(key); field != nil {
		return field
	}
	return nil
}
