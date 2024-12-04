package protocel

import (
	"context"

	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CELContext struct {
	MethodDescriptor protoreflect.MethodDescriptor
	Req              proto.Message
}

func (c *CELContext) ToInput() map[string]any {
	if c == nil {
		return map[string]any{}
	}
	m := map[string]any{
		"req": c.Req,
		"gen": &stubsv1.CELGenerate{},
	}
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
