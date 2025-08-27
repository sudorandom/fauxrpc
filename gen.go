package fauxrpc

import (
	"errors"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	ErrNotFaked = errors.New("no data was faked either because there was no relevant data or the faker was just out of faker juice")
)

type ProtoFaker interface {
	SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error
}

type fauxFaker struct{}

func NewFauxFaker() *fauxFaker {
	return &fauxFaker{}
}

func (f *fauxFaker) SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error {
	return SetDataOnMessage(msg, opts)
}

type multiFaker []ProtoFaker

func NewMultiFaker(fakers []ProtoFaker) multiFaker {
	return multiFaker(fakers)
}

func (f multiFaker) SetDataOnMessage(msg protoreflect.ProtoMessage, opts GenOptions) error {
	for _, faker := range f {
		if err := faker.SetDataOnMessage(msg, opts); err != nil {
			if !errors.Is(err, ErrNotFaked) {
				return err
			}
		} else {
			return nil
		}
	}

	return ErrNotFaked
}

func getFieldConstraints(fd protoreflect.FieldDescriptor, opts GenOptions) (constraints *validate.FieldRules) {
	defer func() {
		if r := recover(); r != nil {
			constraints = nil // Ensure constraints is nil on panic
		}
	}()

	constraints, _ = protovalidate.ResolveFieldRules(fd)

	if opts.extraFieldConstraints != nil {
		if constraints == nil {
			constraints = opts.extraFieldConstraints
		} else {
			// proto.Merge merges the second argument into the first.
			// This means extraFieldConstraints will overwrite base rules where they conflict.
			merged := proto.Clone(constraints).(*validate.FieldRules)
			proto.Merge(merged, opts.extraFieldConstraints)
			constraints = merged
		}
	}

	return constraints
}
