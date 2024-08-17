package fauxrpc

import "google.golang.org/protobuf/reflect/protoreflect"

// GenerateBool returns a fake boolean value given a field descriptor.
func GenerateBool(fd protoreflect.FieldDescriptor) bool {
	// TODO: use protovalidate
	return true
}
