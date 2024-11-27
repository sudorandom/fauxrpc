package stubs

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ActiveIf struct {
	expr    string
	program cel.Program
}

func NewActiveIf(md protoreflect.MethodDescriptor, expr string) (*ActiveIf, error) {
	reqMsg := registry.NewMessage(md.Input()).New()
	env, err := cel.NewEnv(
		cel.Types(reqMsg),
		cel.Variable("req", cel.ObjectType(string(md.Input().FullName()))),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType))
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
