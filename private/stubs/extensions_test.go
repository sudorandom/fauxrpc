package stubs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TestLoadStubsFromFileWithExtension covers a stub that names an extension of
// the message it stubs. protojson only accepts the bracketed key when it can
// resolve the extension, and an extension of a schema loaded at runtime is
// resolvable only through the registry that loaded it.
func TestLoadStubsFromFileWithExtension(t *testing.T) {
	reg, err := registry.NewServiceRegistry()
	require.NoError(t, err)
	require.NoError(t, reg.RegisterFile(loadedFileWithExtension(t)))

	path := filepath.Join(t.TempDir(), "stubs.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
  "stubs": [{
    "id": "traced-event",
    "target": "loaded.v1.EventService/GetEvent",
    "content": {
      "event": {
        "id": "evt_7f3a91",
        "[loaded.v1.trace_id]": "b9c14e2d"
      }
    }
  }]
}`), 0o600))

	database := NewStubDatabase()
	require.NoError(t, LoadStubsFromFile(reg, database, path))

	entries := database.GetStubs()
	require.Len(t, entries, 1)
	msg := entries[0].Message.ProtoReflect()
	event := msg.Get(msg.Descriptor().Fields().ByName("event")).Message()
	assert.Equal(t, "evt_7f3a91", event.Get(event.Descriptor().Fields().ByName("id")).String())

	xt, err := reg.Types().FindExtensionByName("loaded.v1.trace_id")
	require.NoError(t, err)
	require.True(t, event.Has(xt.TypeDescriptor()), "the stub's extension was dropped")
	assert.Equal(t, "b9c14e2d", event.Get(xt.TypeDescriptor()).String())
}

// loadedFileWithExtension builds a schema that is not compiled into this
// binary, which is what fauxrpc actually serves. An extension of a compiled-in
// schema would resolve through the global registry and prove nothing.
func loadedFileWithExtension(t *testing.T) protoreflect.FileDescriptor {
	t.Helper()
	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()
	stringKind := descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()

	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
		Name:    proto.String("loaded/v1/loaded.proto"),
		Package: proto.String("loaded.v1"),
		Syntax:  proto.String("proto2"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Event"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("id"),
					Number: proto.Int32(1),
					Type:   stringKind,
					Label:  optional,
				}},
				ExtensionRange: []*descriptorpb.DescriptorProto_ExtensionRange{{
					Start: proto.Int32(100),
					End:   proto.Int32(200),
				}},
			},
			{
				Name: proto.String("GetEventRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("id"),
					Number: proto.Int32(1),
					Type:   stringKind,
					Label:  optional,
				}},
			},
			{
				Name: proto.String("GetEventResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:     proto.String("event"),
					Number:   proto.Int32(1),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					Label:    optional,
					TypeName: proto.String(".loaded.v1.Event"),
				}},
			},
		},
		Extension: []*descriptorpb.FieldDescriptorProto{{
			Name:     proto.String("trace_id"),
			Number:   proto.Int32(100),
			Type:     stringKind,
			Label:    optional,
			Extendee: proto.String(".loaded.v1.Event"),
		}},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("EventService"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("GetEvent"),
				InputType:  proto.String(".loaded.v1.GetEventRequest"),
				OutputType: proto.String(".loaded.v1.GetEventResponse"),
			}},
		}},
	}}})
	require.NoError(t, err)
	fd, err := files.FindFileByPath("loaded/v1/loaded.proto")
	require.NoError(t, err)
	return fd
}
