package stubs

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ErrNoMatchingStubs = errors.New("no matching stubs")

type stubFaker struct {
	db        StubDatabase
	onlyStubs bool
}

func NewStubFaker(db StubDatabase, onlyStubs bool) *stubFaker {
	return &stubFaker{
		db:        db,
		onlyStubs: onlyStubs,
	}
}

func (f *stubFaker) SetDataOnMessage(pm protoreflect.ProtoMessage, opts fauxrpc.GenOptions) error {
	msg := pm.ProtoReflect()
	desc := msg.Descriptor()
	celCtx := protocel.GetCELContext(opts.GetContext())
	for _, name := range []protoreflect.FullName{celCtx.MethodDescriptor.FullName(), desc.FullName()} {
		for _, group := range f.db.GetStubsPrioritized(name) {
			rand.Shuffle(len(group), func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})

			for _, stub := range group {
				if stub.ActiveIf != nil {
					ok, err := stub.ActiveIf.Eval(context.Background(), celCtx)
					if err != nil {
						return err
					}
					if !ok {
						continue
					}
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
		}
	}

	if f.onlyStubs {
		return ErrNoMatchingStubs
	}
	return nil
}
