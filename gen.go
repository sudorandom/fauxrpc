package fauxrpc

import (
	"errors"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

func (st GenOptions) withExtraFieldConstraints(constraints *validate.FieldConstraints) GenOptions {
	st.extraFieldConstraints = constraints
	return st
}

func getFieldConstraints(fd protoreflect.FieldDescriptor, opts GenOptions) *validate.FieldConstraints {
	if constraints := getResolver().ResolveFieldConstraints(fd); constraints != nil {
		return constraints
	}
	return opts.extraFieldConstraints
}
