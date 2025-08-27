package fauxrpc_test

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Helper to create a field descriptor with specific constraints for testing
type mockFieldDescriptor struct {
	protoreflect.FieldDescriptor
	name         protoreflect.Name
	kind         protoreflect.Kind
	message      protoreflect.MessageDescriptor
	enum         protoreflect.EnumDescriptor
	isMap        bool
	isList       bool
	constraints  *validate.FieldRules
	fieldOptions *descriptorpb.FieldOptions // To store the actual FieldOptions with extension
}

func (m *mockFieldDescriptor) Name() protoreflect.Name {
	return m.name
}

func (m *mockFieldDescriptor) FullName() protoreflect.FullName {
	return protoreflect.FullName(m.name)
}

func (m *mockFieldDescriptor) Kind() protoreflect.Kind {
	return m.kind
}

func (m *mockFieldDescriptor) Message() protoreflect.MessageDescriptor {
	return m.message
}

func (m *mockFieldDescriptor) Enum() protoreflect.EnumDescriptor {
	return m.enum
}

func (m *mockFieldDescriptor) IsMap() bool {
	return m.isMap
}

func (m *mockFieldDescriptor) IsList() bool {
	return m.isList
}

func (m *mockFieldDescriptor) Options() protoreflect.ProtoMessage {
	if m.fieldOptions != nil {
		return m.fieldOptions
	}
	return nil
}

func createFieldDescriptorWithConstraints(baseFd protoreflect.FieldDescriptor, constraints *validate.FieldRules) protoreflect.FieldDescriptor {
	fieldOptions := &descriptorpb.FieldOptions{}
	if constraints != nil {
		proto.SetExtension(fieldOptions, validate.E_Field, constraints)
	}

	return &mockFieldDescriptor{
		FieldDescriptor: baseFd,
		name:            baseFd.Name(),
		kind:            baseFd.Kind(),
		message:         baseFd.Message(),
		enum:            baseFd.Enum(),
		isMap:           baseFd.IsMap(),
		isList:          baseFd.IsList(),
		constraints:     constraints,
		fieldOptions:    fieldOptions,
	}
}
