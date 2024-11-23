package stubs

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Rules struct {
	rules    []string
	programs []cel.Program
}

func NewRules(md protoreflect.MethodDescriptor, strRules []string) (*Rules, error) {
	reqMsg := newMessage(md.Input()).New()
	env, err := cel.NewEnv(
		cel.Types(reqMsg),
		cel.Variable("req", cel.ObjectType(string(md.Input().FullName()))),
		cel.Variable("service", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("procedure", cel.StringType))
	if err != nil {
		return nil, err
	}

	rules := &Rules{
		rules:    strRules,
		programs: make([]cel.Program, len(strRules)),
	}
	for i, strRule := range strRules {
		ast, issues := env.Compile(strRule)
		if issues != nil {
			return nil, issues.Err()
		}
		program, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		rules.programs[i] = program
	}
	return rules, nil
}

func (r *Rules) Eval(ctx context.Context, md protoreflect.MethodDescriptor, req proto.Message) (bool, error) {
	for _, p := range r.programs {
		val, _, err := p.ContextEval(ctx, map[string]any{
			"req":       req,
			"service":   string(md.Parent().FullName()),
			"method":    string(md.Name()),
			"procedure": string(md.FullName()),
		})
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
	}
	return true, nil
}

func (r *Rules) GetStrings() []string {
	return r.rules
}
