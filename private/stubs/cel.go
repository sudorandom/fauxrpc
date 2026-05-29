package stubs

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/sudorandom/fauxrpc/celfakeit"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ActiveIf struct {
	expr    string
	program cel.Program
}

func NewActiveIf(md protoreflect.MethodDescriptor, expr string) (*ActiveIf, error) {
	reqMsg := registry.NewMessage(md.Input()).New()
	env, err := cel.NewEnv(
		celfakeit.Configure(),
		ext.Encoders(),
		cel.Types(reqMsg),
		cel.Variable("req", cel.ObjectType(string(md.Input().FullName()))),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType),
		cel.Variable("now", cel.TimestampType),
	)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expr)
	if issues != nil {
		return nil, issues.Err()
	}
	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("output type should be bool, actual=%s", ast.OutputType())
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return &ActiveIf{
		expr:    expr,
		program: program,
	}, nil
}

func (r *ActiveIf) Eval(ctx context.Context, celCtx *protocel.CELContext) (bool, error) {
	input := celCtx.ToInput()
	val, _, err := r.program.ContextEval(ctx, input)
	if err != nil {
		return false, err
	}
	switch t := val.Value().(type) {
	case bool:
		if !t {
			return false, nil
		}
	default:
		return false, fmt.Errorf("unexpected return type from CEL expr (%T): %+v", t, val)
	}
	return true, nil
}

func (r *ActiveIf) GetString() string {
	return r.expr
}

func ActiveIfFromProto(msg proto.Message) string {
	if msg == nil {
		return ""
	}
	var conds []string
	msg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		if fd.IsMap() || fd.IsList() || fd.Kind() == protoreflect.MessageKind {
			return true
		}
		valStr := formatCELValue(val, fd)
		if valStr != "" {
			conds = append(conds, fmt.Sprintf("req.%s == %s", fd.Name(), valStr))
		}
		return true
	})
	if len(conds) == 0 {
		return ""
	}
	return strings.Join(conds, " && ")
}

func formatCELValue(val protoreflect.Value, fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return fmt.Sprintf("%t", val.Bool())
	case protoreflect.StringKind:
		return fmt.Sprintf("%q", val.String())
	case protoreflect.BytesKind:
		return fmt.Sprintf("b%q", val.Bytes())
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return fmt.Sprintf("%d", val.Int())
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		return fmt.Sprintf("%du", val.Uint())
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return fmt.Sprintf("%g", val.Float())
	case protoreflect.EnumKind:
		return fmt.Sprintf("%d", val.Enum())
	default:
		return ""
	}
}
