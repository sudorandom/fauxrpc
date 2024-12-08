package protocel

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type CELMessage interface {
	NewMessage(ctx context.Context) (proto.Message, error)
	SetDataOnMessage(ctx context.Context, msg protoreflect.ProtoMessage) error
}

var validTopLevelTypes = []*types.Type{
	types.DynType,
	cel.MapType(types.StringType, types.DynType),
	cel.MapType(types.StringType, cel.MapType(types.StringType, types.StringType)),
	cel.MapType(types.StringType, cel.MapType(types.StringType, types.DynType)),
}

var _ CELMessage = (*protocel)(nil)

type protocel struct {
	messageDescriptor protoreflect.MessageDescriptor
	program           cel.Program
}

func New(files *protoregistry.Files, md protoreflect.MessageDescriptor, celString string) (*protocel, error) {
	env, err := newEnv(files)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(celString)
	if issues != nil {
		return nil, issues.Err()
	}
	if !isCELType(ast.OutputType(), validTopLevelTypes...) {
		return nil, fmt.Errorf("%s: unexpected type '%s'; wanted one of: %v", md.FullName(), ast.OutputType(), validTopLevelTypes)
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return &protocel{
		messageDescriptor: md,
		program:           program,
	}, nil
}

// NewMessage implements CELMessage.
func (p *protocel) NewMessage(ctx context.Context) (protoreflect.ProtoMessage, error) {
	msg := registry.NewMessage(p.messageDescriptor).Interface()
	if err := p.SetDataOnMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// SetDataOnMessage implements CELMessage.
func (p *protocel) SetDataOnMessage(ctx context.Context, pmsg protoreflect.ProtoMessage) error {
	input := GetCELContext(ctx).ToInput()
	val, _, err := p.program.Eval(input)
	if err != nil {
		return fmt.Errorf("cel eval: %w", err)
	}
	msg := pmsg.ProtoReflect()
	switch tval := val.Value().(type) {
	case map[ref.Val]ref.Val:
		return p.setFieldsOnMsg(msg, tval)
	case *stubsv1.CELGenerate:
		return fauxrpc.SetDataOnMessage(pmsg, fauxrpc.GenOptions{MaxDepth: 5})
	case proto.Message:
		outMsg := tval.ProtoReflect()
		if msg.Descriptor() != outMsg.Descriptor() {
			got, want := outMsg.Descriptor().FullName(), outMsg.Descriptor().FullName()
			return fmt.Errorf("descriptor mismatch: %v != %v", got, want)
		}

		proto.Merge(pmsg, tval)
		return nil
	default:
		return fmt.Errorf("%s: unhandled type: %T", msg.Descriptor().FullName(), val.Value())
	}
}

func (p *protocel) setFieldsOnMsg(msg protoreflect.Message, fields map[ref.Val]ref.Val) error {
	desc := msg.Descriptor()
	msgFields := desc.Fields()
	for refKey, refVal := range fields {
		key := refKey.ConvertToType(types.StringType).Value().(string)
		fd := getFieldFromName(msgFields, key)
		if fd == nil {
			return fmt.Errorf("%s: field does not exist: %s", desc.FullName(), key)
		}
		val := refVal.Value()
		switch tval := val.(type) {
		case nil:
			return nil
		case map[ref.Val]ref.Val:
			if fd.IsMap() {
				if err := p.setMapField(msg, fd, tval); err != nil {
					return err
				}
			} else {
				nested := registry.NewMessage(fd.Message()).New()
				if err := p.setFieldsOnMsg(nested, tval); err != nil {
					return err
				}
				msg.Set(fd, protoreflect.ValueOfMessage(nested))
			}
		case []ref.Val:
			if err := p.setRepeatedField(msg, fd, tval); err != nil {
				return err
			}
		case *stubsv1.CELGenerate:
			if val := fauxrpc.FieldValue(fd, fauxrpc.GenOptions{
				MaxDepth: 5,
			}); val != nil {
				msg.Set(fd, *val)
			}
		default:
			value, err := p.celToValue(fd, val)
			if err != nil {
				return err
			}
			msg.Set(fd, value)
		}
	}
	return nil
}

func (p *protocel) setMapField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, fields map[ref.Val]ref.Val) error {
	m := msg.NewField(fd).Map()
	for k, v := range fields {
		val, err := p.celToValue(fd.MapValue(), v.Value())
		if err != nil {
			return err
		}
		m.Set(protoreflect.ValueOf(k.Value()).MapKey(), val)
	}
	msg.Set(fd, protoreflect.ValueOfMap(m))
	return nil
}

func (p *protocel) setRepeatedField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, vals []ref.Val) error {
	if fd.Cardinality() != protoreflect.Repeated {
		return fmt.Errorf("%s: list returned for a non-repeated field: %s", msg.Descriptor().FullName(), fd.Name())
	}
	list := msg.NewField(fd).List()
	switch fd.Kind() {
	case protoreflect.MessageKind:
		nested := registry.NewMessage(fd.Message()).Interface()
		for _, val := range vals {
			mapVal, ok := val.Value().(map[ref.Val]ref.Val)
			if !ok {
				return fmt.Errorf("%s: unhandled type: %T", msg.Descriptor().FullName(), val.Value())
			}
			if err := p.setFieldsOnMsg(nested.ProtoReflect(), mapVal); err != nil {
				return err
			}
		}
		list.Append(protoreflect.ValueOfMessage(nested.ProtoReflect()))
	default:
		for _, val := range vals {
			value, err := p.celToValue(fd, val.Value())
			if err != nil {
				return err
			}
			list.Append(value)
		}
	}
	msg.Set(fd, protoreflect.ValueOfList(list))
	return nil
}

func (p *protocel) celToValue(fd protoreflect.FieldDescriptor, val any) (protoreflect.Value, error) {
	switch tv := val.(type) {
	case map[ref.Val]ref.Val:
		nested := registry.NewMessage(fd.Message()).New()
		if err := p.setFieldsOnMsg(nested, tv); err != nil {
			return protoreflect.ValueOf(nil), err
		}
		return protoreflect.ValueOfMessage(nested), nil
	case *stubsv1.CELGenerate:
		if val := fauxrpc.FieldValue(fd, fauxrpc.GenOptions{
			MaxDepth: 5,
		}); val != nil {
			return *val, nil
		} else {
			return protoreflect.ValueOf(nil), nil
		}
	case proto.Message:
		return protoreflect.ValueOfMessage(tv.ProtoReflect()), nil
	default:
		switch fd.Kind() {
		case protoreflect.EnumKind:
			switch t := val.(type) {
			case int64:
				return protoreflect.ValueOfEnum(protoreflect.EnumNumber(t)), nil
			case uint64:
				return protoreflect.ValueOfEnum(protoreflect.EnumNumber(t)), nil
			case string:
				v := fd.Enum().Values().ByName(protoreflect.Name(t))
				if v == nil {
					return protoreflect.ValueOf(nil), fmt.Errorf("unknown enum value: '%s'", t)
				}
				return protoreflect.ValueOfEnum(v.Number()), nil
			}
		}
	}
	return protoreflect.ValueOf(val), nil
}

func isCELType(t *types.Type, targets ...*types.Type) bool {
	for _, target := range targets {
		if target.IsAssignableRuntimeType(t) {
			return true
		}
	}
	return false
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

func newEnv(files *protoregistry.Files) (*cel.Env, error) {
	return cel.NewEnv(
		cel.TypeDescs(files),
		cel.Variable("req", cel.DynType),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType),
		cel.Variable("gen", cel.DynType),
		cel.Types(&stubsv1.CELGenerate{}),
	)
}
