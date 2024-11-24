package protocel

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DynamicMessage interface {
	NewMessage(opts fauxrpc.GenOptions) (proto.Message, error)
	SetDataOnMessage(msg protoreflect.ProtoMessage, opts fauxrpc.GenOptions) error
}

var _ DynamicMessage = (*dynamicMessage)(nil)

type dynamicMessage struct {
	messageDescriptor protoreflect.MessageDescriptor
	fields            map[protoreflect.FieldDescriptor]cel.Program
	nested            map[protoreflect.FieldDescriptor]*dynamicMessage
	repeatedMsg       map[protoreflect.FieldDescriptor][]*dynamicMessage
	repeatedScalar    map[protoreflect.FieldDescriptor][]cel.Program
	mapsMsg           map[protoreflect.FieldDescriptor]map[cel.Program]*dynamicMessage
	mapsScalar        map[protoreflect.FieldDescriptor]map[cel.Program]cel.Program
}

func NewDynamicMessage(md protoreflect.MessageDescriptor, fields map[string]Node) (*dynamicMessage, error) {
	celFields := map[protoreflect.FieldDescriptor]cel.Program{}
	nested := map[protoreflect.FieldDescriptor]*dynamicMessage{}
	repeatedMsg := map[protoreflect.FieldDescriptor][]*dynamicMessage{}
	repeatedScalar := map[protoreflect.FieldDescriptor][]cel.Program{}
	mapsMsg := map[protoreflect.FieldDescriptor]map[cel.Program]*dynamicMessage{}
	mapsScalar := map[protoreflect.FieldDescriptor]map[cel.Program]cel.Program{}
	for key, node := range fields {
		field := getFieldFromName(md.Fields(), key)
		if field == nil {
			return nil, fmt.Errorf("field %s not found on %s", key, md.FullName())
		}
		switch node.Kind() {
		case CELKind:
			celnode := node.(nodeCEL)
			program, err := compileExpr(md, field, string(celnode))
			if err != nil {
				return nil, err
			}

			celFields[field] = program
		case MessageKind:
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is expected to be a message but was %s", key, field.Kind())
			}
			messageNode := node.(nodeMessage)
			nestedNode, err := NewDynamicMessage(field.Message(), messageNode)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			nested[field] = nestedNode
		case RepeatedKind:
			if !field.IsList() {
				return nil, fmt.Errorf("field %s is expected to be a list but was not", key)
			}
			if field.Kind() == protoreflect.MessageKind {
				repeated := node.(nodeRepeated)
				for _, node := range repeated {
					messageNode := node.(nodeMessage)
					nestedNode, err := NewDynamicMessage(field.Message(), messageNode)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", key, err)
					}
					repeatedMsg[field] = append(repeatedMsg[field], nestedNode)
				}
			} else {
				repeated := node.(nodeRepeated)
				for _, node := range repeated {
					celnode := node.(nodeCEL)
					program, err := compileExpr(md, field, string(celnode))
					if err != nil {
						return nil, err
					}
					repeatedScalar[field] = append(repeatedScalar[field], program)
				}
			}
		case MapKind:
			if !field.IsMap() {
				return nil, fmt.Errorf("field %s is expected to be a map but was not", key)
			}

			nMap := node.(nodeMap)
			for k, v := range nMap {
				if k.Kind() != CELKind {
					return nil, fmt.Errorf("key %s field for maps is expected to be a CEL expression but was %v", key, k.Kind())
				}
				keyProgram, err := compileExpr(md, field.MapKey(), string(k.(nodeCEL)))
				if err != nil {
					return nil, err
				}

				switch v.Kind() {
				case CELKind:
					valProgram, err := compileExpr(md, field.MapValue(), string(v.(nodeCEL)))
					if err != nil {
						return nil, err
					}
					if _, ok := mapsScalar[field]; !ok {
						mapsScalar[field] = map[cel.Program]cel.Program{}
					}
					mapsScalar[field][keyProgram] = valProgram
				case MessageKind:
					valNode, err := NewDynamicMessage(field.MapValue().Message(), v.(nodeMessage))
					if err != nil {
						return nil, fmt.Errorf("%s: %w", key, err)
					}
					if _, ok := mapsMsg[field]; !ok {
						mapsMsg[field] = map[cel.Program]*dynamicMessage{}
					}
					mapsMsg[field][keyProgram] = valNode
				}
			}

		default:
			return nil, fmt.Errorf("%s: unknown node kind: %v", key, node.Kind())
		}
	}

	return &dynamicMessage{
		messageDescriptor: md,
		fields:            celFields,
		nested:            nested,
		repeatedMsg:       repeatedMsg,
		repeatedScalar:    repeatedScalar,
		mapsMsg:           mapsMsg,
		mapsScalar:        mapsScalar,
	}, nil
}

// NewMessage implements DynamicMessage.
func (d *dynamicMessage) NewMessage(opts fauxrpc.GenOptions) (protoreflect.ProtoMessage, error) {
	msg := registry.NewMessage(d.messageDescriptor).Interface()
	if err := d.SetDataOnMessage(msg, opts); err != nil {
		return nil, err
	}
	return msg, nil
}

// SetDataOnMessage implements DynamicMessage.
func (d *dynamicMessage) SetDataOnMessage(msg protoreflect.ProtoMessage, opts fauxrpc.GenOptions) error {
	// TODO: this input should come from GenOptions (or some other context object instead
	input := map[string]any{}
	for field, program := range d.fields {
		val, err := evalCEL(field, program, input)
		if err != nil {
			return err
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOf(val))
	}
	for field, dynmsg := range d.nested {
		nestedMsg := registry.NewMessage(field.Message()).Interface()
		if err := dynmsg.SetDataOnMessage(nestedMsg, opts); err != nil {
			return err
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOfMessage(nestedMsg.ProtoReflect()))
	}
	for field, dynmsgs := range d.repeatedMsg {
		list := msg.ProtoReflect().NewField(field).List()
		for _, dynmsg := range dynmsgs {
			nestedMsg := registry.NewMessage(field.Message()).Interface()
			if err := dynmsg.SetDataOnMessage(nestedMsg, opts); err != nil {
				return err
			}
			list.Append(protoreflect.ValueOfMessage(nestedMsg.ProtoReflect()))
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOf(list))
	}
	for field, scalarMsgs := range d.repeatedScalar {
		list := msg.ProtoReflect().NewField(field).List()
		for _, program := range scalarMsgs {
			val, err := evalCEL(field, program, input)
			if err != nil {
				return err
			}

			list.Append(protoreflect.ValueOf(val))
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOfList(list))
	}
	for field, dynMsgMap := range d.mapsMsg {
		m := msg.ProtoReflect().NewField(field).Map()
		for kProg, dynMsg := range dynMsgMap {
			key, _, err := kProg.Eval(input)
			if err != nil {
				return err
			}
			nestedMsg := registry.NewMessage(field.MapValue().Message()).Interface()
			if err := dynMsg.SetDataOnMessage(nestedMsg, opts); err != nil {
				return err
			}

			m.Set(protoreflect.ValueOf(key.Value()).MapKey(), protoreflect.ValueOf(nestedMsg.ProtoReflect()))
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOfMap(m))
	}
	for field, mapScalar := range d.mapsScalar {
		m := msg.ProtoReflect().NewField(field).Map()
		for kProg, vProg := range mapScalar {
			key, _, err := kProg.Eval(input)
			if err != nil {
				return err
			}
			val, _, err := vProg.Eval(input)
			if err != nil {
				return err
			}

			m.Set(protoreflect.ValueOf(key.Value()).MapKey(), protoreflect.ValueOf(val.Value()))
		}
		msg.ProtoReflect().Set(field, protoreflect.ValueOfMap(m))
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

func compileExpr(md protoreflect.MessageDescriptor, fd protoreflect.FieldDescriptor, expr string) (cel.Program, error) {
	env, err := cel.NewEnv(
		cel.Variable("req", cel.ObjectType(string(md.FullName()))),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType),
		cel.Function("gen_bool",
			cel.Overload("gen_bool_noargs", []*cel.Type{}, types.BoolType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Bool(fauxrpc.Bool(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_bytes",
			cel.Overload("gen_bytes_noargs", []*cel.Type{}, types.BytesType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Bytes(fauxrpc.Bytes(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_enum",
			cel.Overload("gen_enum_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.Enum(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_fixed32",
			cel.Overload("gen_fixed32_noargs", []*cel.Type{}, types.UintType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Uint(fauxrpc.Fixed32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_fixed64",
			cel.Overload("gen_fixed64_noargs", []*cel.Type{}, types.UintType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Uint(fauxrpc.Fixed64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_float32",
			cel.Overload("gen_float32_noargs", []*cel.Type{}, types.DoubleType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Double(fauxrpc.Float32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_float64",
			cel.Overload("gen_float64_noargs", []*cel.Type{}, types.DoubleType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Double(fauxrpc.Float64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_int32",
			cel.Overload("gen_int32_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.Int32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_int64",
			cel.Overload("gen_int64_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.Int64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sfixed32",
			cel.Overload("gen_sfixed32_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.SFixed32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sfixed64",
			cel.Overload("gen_sfixed64_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.SFixed64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sint32",
			cel.Overload("gen_sint32_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.SInt32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_sint64",
			cel.Overload("gen_sint64_noargs", []*cel.Type{}, types.IntType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Int(fauxrpc.SInt64(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_string",
			cel.Overload("gen_string_noargs", []*cel.Type{}, types.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.String(fauxrpc.String(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_uint32",
			cel.Overload("gen_uint32_noargs", []*cel.Type{}, types.UintType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Uint(fauxrpc.UInt32(fd, fauxrpc.GenOptions{}))
				}),
			),
		),
		cel.Function("gen_uint64",
			cel.Overload("gen_uint64_noargs", []*cel.Type{}, types.UintType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Uint(fauxrpc.UInt64(fd, fauxrpc.GenOptions{}))
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
	if err := checkCelType(fd, ast.OutputType()); err != nil {
		return nil, fmt.Errorf("%s: %w", fd.Name(), err)
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func evalCEL(field protoreflect.FieldDescriptor, program cel.Program, input map[string]any) (any, error) {
	val, _, err := program.Eval(input)
	if err != nil {
		return nil, err
	}
	anyVal := val.Value()
	switch field.Kind() {
	case protoreflect.EnumKind:
		switch t := anyVal.(type) {
		case int64:
			anyVal = protoreflect.EnumNumber(t)
		case uint64:
			anyVal = protoreflect.EnumNumber(t)
		case string:
			anyVal = field.Enum().Values().ByName(protoreflect.Name(t))
			if anyVal == nil {
				return nil, fmt.Errorf("unknown enum value: '%s'", t)
			}
		}
	}
	return anyVal, nil
}
