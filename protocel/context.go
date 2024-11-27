package protocel

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CELContext struct {
	MethodDescriptor protoreflect.MethodDescriptor
	Input            proto.Message
}

func (c *CELContext) ToInput() map[string]any {
	if c == nil {
		return map[string]any{}
	}
	return map[string]any{
		"req":       c.Input,
		"service":   string(c.MethodDescriptor.Parent().FullName()),
		"method":    string(c.MethodDescriptor.Name()),
		"procedure": string(c.MethodDescriptor.FullName()),
	}
}

type celCtxKeyType string

const celCtxKey celCtxKeyType = "celCtx"

func WithCELContext(ctx context.Context, celCtx CELContext) context.Context {
	return context.WithValue(ctx, celCtx, &celCtx)
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
