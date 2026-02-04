package stubs

import (
	"context"
	"math/rand/v2"

	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type StubFaker struct {
	db StubDatabase
}

func NewStubFaker(db StubDatabase) *StubFaker {
	return &StubFaker{db: db}
}

func (f *StubFaker) FindStub(ctx context.Context, celCtx *protocel.CELContext, desc protoreflect.MessageDescriptor) (*StubEntry, error) {
	for _, name := range []protoreflect.FullName{celCtx.MethodDescriptor.FullName(), desc.FullName()} {
		for _, group := range f.db.GetStubsPrioritized(name) {
			rand.Shuffle(len(group), func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})

			for _, stub := range group {
				if stub.ActiveIf != nil {
					ok, err := stub.ActiveIf.Eval(ctx, celCtx)
					if err != nil {
						return nil, err
					}
					if !ok {
						continue
					}
				}
				return &stub, nil
			}
		}
	}
	return nil, nil
}

func (f *StubFaker) SetDataOnMessage(pm protoreflect.ProtoMessage, opts fauxrpc.GenOptions) error {
	msg := pm.ProtoReflect()
	desc := msg.Descriptor()
	celCtx := protocel.GetCELContext(opts.GetContext())

	stub, err := f.FindStub(opts.GetContext(), celCtx, desc)
	if err != nil {
		return err
	}
	if stub == nil {
		return fauxrpc.ErrNotFaked
	}

	if opts.StubRecorder != nil {
		opts.StubRecorder(stub.Key)
	}
	if stub.Error != nil {
		return stub.Error
	}
	if stub.Message != nil {
		fields := desc.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if stub.Message.ProtoReflect().Has(field) {
				msg.Set(field, stub.Message.ProtoReflect().Get(field))
			}
		}
	}
	if stub.CELMessage != nil {
		if err := stub.CELMessage.SetDataOnMessage(opts.GetContext(), msg.Interface()); err != nil {
			return err
		}
	}
	return nil
}
