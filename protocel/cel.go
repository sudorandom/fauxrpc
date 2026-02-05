package protocel

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/google/cel-go/ext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/celfakeit"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/registry"
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

type Compiler struct {
	env          *cel.Env
	programCache map[string]cel.Program
	mu           sync.RWMutex
}

func NewCompiler(files *protoregistry.Files) (*Compiler, error) {
	env, err := newEnv(files)
	if err != nil {
		return nil, err
	}
	return &Compiler{
		env:          env,
		programCache: make(map[string]cel.Program),
	}, nil
}

func (c *Compiler) Compile(md protoreflect.MessageDescriptor, celString string) (CELMessage, error) {
	c.mu.RLock()
	program, ok := c.programCache[celString]
	c.mu.RUnlock()

	if !ok {
		ast, issues := c.env.Compile(celString)
		if issues != nil {
			return nil, issues.Err()
		}
		if !isCELType(ast.OutputType(), validTopLevelTypes...) {
			return nil, fmt.Errorf("%s: unexpected type '%s'; wanted one of: %v", md.FullName(), ast.OutputType(), validTopLevelTypes)
		}

		var err error
		program, err = c.env.Program(ast)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.programCache[celString] = program
		c.mu.Unlock()
	}

	return &protocel{
		messageDescriptor: md,
		program:           program,
	}, nil
}

func New(files *protoregistry.Files, md protoreflect.MessageDescriptor, celString string) (CELMessage, error) {
	c, err := NewCompiler(files)
	if err != nil {
		return nil, err
	}
	return c.Compile(md, celString)
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
func (p *protocel) SetDataOnMessage(ctx context.Context, pmsg protoreflect.ProtoMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch tt := r.(type) {
			case string:
				err = errors.New(tt)
			case error:
				err = tt
			default:
				err = fmt.Errorf("%+v", tt)
			}
		}
	}()
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
		if _, ok := val.(structpb.NullValue); ok {
			continue
		}
		if val == nil {
			continue
		}
		switch tval := val.(type) {
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
		default:
			if fd.Cardinality() == protoreflect.Repeated {
				list := msg.NewField(fd).List()
				value, err := p.celToValue(fd, val)
				if err != nil {
					return err
				}
				list.Append(value)
				msg.Set(fd, protoreflect.ValueOfList(list))
			} else {
				value, err := p.celToValue(fd, val)
				if err != nil {
					return err
				}
				msg.Set(fd, value)
			}
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
		for _, val := range vals {
			mapVal, ok := val.Value().(map[ref.Val]ref.Val)
			if !ok {
				return fmt.Errorf("%s: unhandled type: %T", msg.Descriptor().FullName(), val.Value())
			}
			nested := registry.NewMessage(fd.Message()).Interface()
			if err := p.setFieldsOnMsg(nested.ProtoReflect(), mapVal); err != nil {
				return err
			}
			list.Append(protoreflect.ValueOfMessage(nested.ProtoReflect()))
		}
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
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			switch t := val.(type) {
			case int64:
				return protoreflect.ValueOfInt32(int32(t)), nil
			case uint64:
				return protoreflect.ValueOfInt32(int32(t)), nil
			case int32:
				return protoreflect.ValueOfInt32(t), nil
			}
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			switch t := val.(type) {
			case int64:
				return protoreflect.ValueOfUint32(uint32(t)), nil
			case uint64:
				return protoreflect.ValueOfUint32(uint32(t)), nil
			case uint32:
				return protoreflect.ValueOfUint32(t), nil
			}
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			switch t := val.(type) {
			case int64:
				return protoreflect.ValueOfInt64(t), nil
			case uint64:
				return protoreflect.ValueOfInt64(int64(t)), nil
			case int32:
				return protoreflect.ValueOfInt64(int64(t)), nil
			}
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			switch t := val.(type) {
			case int64:
				return protoreflect.ValueOfUint64(uint64(t)), nil
			case uint64:
				return protoreflect.ValueOfUint64(t), nil
			case uint32:
				return protoreflect.ValueOfUint64(uint64(t)), nil
			}
		case protoreflect.FloatKind:
			switch t := val.(type) {
			case float64:
				return protoreflect.ValueOfFloat32(float32(t)), nil
			case float32:
				return protoreflect.ValueOfFloat32(t), nil
			case int64:
				return protoreflect.ValueOfFloat32(float32(t)), nil
			case uint64:
				return protoreflect.ValueOfFloat32(float32(t)), nil
			}
		case protoreflect.DoubleKind:
			switch t := val.(type) {
			case float64:
				return protoreflect.ValueOfFloat64(t), nil
			case float32:
				return protoreflect.ValueOfFloat64(float64(t)), nil
			case int64:
				return protoreflect.ValueOfFloat64(float64(t)), nil
			case uint64:
				return protoreflect.ValueOfFloat64(float64(t)), nil
			}
		case protoreflect.BytesKind:
			switch t := val.(type) {
			case []byte:
				return protoreflect.ValueOfBytes(t), nil
			case string:
				// Attempt to base64 decode the string, if it fails, use the raw string bytes
				decoded, err := base64.StdEncoding.DecodeString(t)
				if err == nil {
					return protoreflect.ValueOfBytes(decoded), nil
				}
				return protoreflect.ValueOfBytes([]byte(t)), nil
			}
		case protoreflect.MessageKind:
			if fd.Message().FullName() == "google.protobuf.Timestamp" {
				switch t := val.(type) {
				case time.Time:
					ts := timestamppb.New(t)
					return protoreflect.ValueOfMessage(ts.ProtoReflect()), nil
				case string:
					parsedTime, err := time.Parse(time.RFC3339Nano, t)
					if err != nil {
						return protoreflect.ValueOf(nil), fmt.Errorf("failed to parse timestamp: %w", err)
					}
					ts := timestamppb.New(parsedTime)
					return protoreflect.ValueOfMessage(ts.ProtoReflect()), nil
				}
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
		celfakeit.Configure(),
		ext.Encoders(),
		cel.TypeDescs(files),
		cel.Variable("req", cel.DynType),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType),
		cel.Variable("gen", cel.DynType),
		cel.Variable("faker", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		cel.Types(&stubsv1.CELGenerate{}),
	)
}
