package stubs

import (
	"context"
	"time"

	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Sender func(proto.Message) error
type FallbackGenerator func(proto.Message) error

func ExecuteStream(
	ctx context.Context,
	stream *StreamEntry,
	msgDesc protoreflect.MessageDescriptor,
	celCtx *protocel.CELContext,
	sender Sender,
	fallbackGenerator FallbackGenerator,
) error {
	startTime := time.Now()

	for stream.DoneAfter == 0 || time.Since(startTime) <= stream.DoneAfter {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		for _, item := range stream.Items {
			if stream.DoneAfter > 0 && time.Since(startTime) > stream.DoneAfter {
				break
			}
			if item.Delay > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(item.Delay):
				}
			}

			if item.Error != nil {
				return item.Error
			}

			out := registry.NewMessage(msgDesc).Interface()
			modified := false

			if item.Message != nil {
				proto.Merge(out, item.Message)
				modified = true
			}

			if item.CELMessage != nil {
				ctxWithCEL := protocel.WithCELContext(ctx, celCtx)
				if err := item.CELMessage.SetDataOnMessage(ctxWithCEL, out); err != nil {
					return err
				}
				modified = true
			}

			if !modified && fallbackGenerator != nil {
				if err := fallbackGenerator(out); err != nil {
					return err
				}
			}

			if err := sender(out); err != nil {
				return err
			}
		}

		if !stream.Repeated {
			break
		}
	}
	return nil
}
