package protocel

import (
	"context"
	"time"

	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CELContext struct {
	MethodDescriptor protoreflect.MethodDescriptor
	Req              proto.Message
}

func (c *CELContext) ToInput() map[string]any {
	m := map[string]any{
		"gen": &stubsv1.CELGenerate{},
		"now": time.Now(),
	}

	if c == nil {
		return m
	}

	m["req"] = c.Req
	if c.MethodDescriptor != nil {
		m["service"] = string(c.MethodDescriptor.Parent().FullName())
		m["method"] = string(c.MethodDescriptor.Name())
		m["procedure"] = string(c.MethodDescriptor.FullName())

	}
	return m
}

type celCtxKeyType string

const celCtxKey celCtxKeyType = "celCtx"

func WithCELContext(ctx context.Context, celCtx *CELContext) context.Context {
	ctx = context.WithValue(ctx, celCtxKey, celCtx)
	return ctx
}

func GetCELContext(ctx context.Context) *CELContext {
	if ctx == nil {
		return nil
	}
	celCtx, ok := ctx.Value(celCtxKey).(*CELContext)
	if !ok {
		return nil
	}
	return celCtx
}
