package registry

import (
	"errors"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Resolver resolves the types that JSON names but a message descriptor cannot
// reach on its own: the extensions of a message, written as "[full.name]", and
// the message packed inside a google.protobuf.Any.
//
// Without one, protojson rejects any JSON naming an extension as an unknown
// field, so a stub could not carry a value that the server itself generates.
type Resolver struct {
	types *protoregistry.Types
}

// Resolver returns a resolver over the loaded schema, falling back to the types
// compiled into this binary for anything the schema does not declare.
func (r *serviceRegistry) Resolver() *Resolver {
	return &Resolver{types: r.Types()}
}

func (r *Resolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	xt, err := r.types.FindExtensionByName(field)
	if isNotFound(err) {
		return protoregistry.GlobalTypes.FindExtensionByName(field)
	}
	return xt, err
}

func (r *Resolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	xt, err := r.types.FindExtensionByNumber(message, field)
	if isNotFound(err) {
		return protoregistry.GlobalTypes.FindExtensionByNumber(message, field)
	}
	return xt, err
}

func (r *Resolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	mt, err := r.types.FindMessageByName(message)
	if isNotFound(err) {
		return protoregistry.GlobalTypes.FindMessageByName(message)
	}
	return mt, err
}

func (r *Resolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	mt, err := r.types.FindMessageByURL(url)
	if isNotFound(err) {
		return protoregistry.GlobalTypes.FindMessageByURL(url)
	}
	return mt, err
}

func isNotFound(err error) bool {
	return err != nil && errors.Is(err, protoregistry.NotFound)
}
