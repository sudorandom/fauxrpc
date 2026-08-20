package fauxrpc_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestExtensions(t *testing.T) {
	opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(1), Extensions: protoregistry.GlobalTypes}

	t.Run("no resolver, no extensions", func(t *testing.T) {
		// The descriptor of a message cannot name the extensions declared
		// against it, so with nothing to ask, there is nothing to generate.
		for range 50 {
			msg := &testv1.Event{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{MaxDepth: 5}))
			assert.NotEmpty(t, msg.GetId())
			assert.False(t, proto.HasExtension(msg, testv1.E_TraceId))
			assert.False(t, proto.HasExtension(msg, testv1.E_Sampled))
		}
	})

	t.Run("sets the extensions of the schema", func(t *testing.T) {
		// Each extension gets its own roll, so every one of them shows up
		// eventually and none of them shows up every time.
		set := map[string]int{}
		const runs = 200
		for range runs {
			msg := &testv1.Event{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, opts))
			for _, xt := range []protoreflect.ExtensionType{testv1.E_TraceId, testv1.E_Sampled, testv1.E_RetryCount, testv1.E_Labels, testv1.E_Source} {
				if proto.HasExtension(msg, xt) {
					set[string(xt.TypeDescriptor().FullName())]++
				}
			}
		}
		for _, name := range []string{"test.v1.trace_id", "test.v1.sampled", "test.v1.retry_count", "test.v1.labels", "test.v1.source"} {
			assert.Greater(t, set[name], 0, "%s was never set", name)
			assert.Less(t, set[name], runs, "%s was set every time", name)
		}
	})

	t.Run("values match their type", func(t *testing.T) {
		for range 200 {
			msg := &testv1.Event{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, opts))
			if proto.HasExtension(msg, testv1.E_Labels) {
				assert.NotEmpty(t, proto.GetExtension(msg, testv1.E_Labels).([]string))
			}
			if proto.HasExtension(msg, testv1.E_Source) {
				source, ok := proto.GetExtension(msg, testv1.E_Source).(*testv1.EventSource)
				require.True(t, ok)
				assert.NotEmpty(t, source.GetService())
			}
			// A round trip through the wire proves the values sit on the field
			// numbers the extensions claim.
			wire, err := proto.Marshal(msg)
			require.NoError(t, err)
			back := &testv1.Event{}
			require.NoError(t, proto.Unmarshal(wire, back))
			assert.True(t, proto.Equal(msg, back))
		}
	})

	t.Run("reaches nested messages", func(t *testing.T) {
		var sawNested bool
		for range 50 {
			msg := &testv1.EventBatch{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, opts))
			for _, event := range msg.GetEvents() {
				if proto.HasExtension(event, testv1.E_TraceId) {
					sawNested = true
				}
			}
		}
		assert.True(t, sawNested, "extensions were never set on a nested message")
	})

	t.Run("JSON round trips through a resolver", func(t *testing.T) {
		for range 50 {
			msg := &testv1.Event{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, opts))
			jsonBytes, err := protojson.Marshal(msg)
			require.NoError(t, err)
			back := &testv1.Event{}
			require.NoError(t, protojson.Unmarshal(jsonBytes, back))
			assert.True(t, proto.Equal(msg, back), "%s", jsonBytes)
		}
	})

	t.Run("descriptor options are left alone", func(t *testing.T) {
		// Every custom option in a schema is an extension of one of these, and
		// generating them would describe a schema that does not exist.
		for range 50 {
			fieldOpts := &descriptorpb.FieldOptions{}
			require.NoError(t, fauxrpc.SetDataOnMessage(fieldOpts, fauxrpc.GenOptions{MaxDepth: 5, Extensions: protoregistry.GlobalTypes}))
			count := 0
			fieldOpts.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
				if fd.IsExtension() {
					count++
				}
				return true
			})
			assert.Zero(t, count, "%d extensions were set on FieldOptions", count)
		}
	})
}

// TestExtensionsFromLoadedSchema covers the case the server actually runs: a
// schema loaded at runtime, where neither the message nor its extensions have a
// Go type and both come from the registry as dynamic ones.
func TestExtensionsFromLoadedSchema(t *testing.T) {
	reg, err := registry.NewServiceRegistry()
	require.NoError(t, err)

	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
		Name:    proto.String("loaded/v1/loaded.proto"),
		Package: proto.String("loaded.v1"),
		Syntax:  proto.String("proto2"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Event"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("id"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			}},
			ExtensionRange: []*descriptorpb.DescriptorProto_ExtensionRange{{
				Start: proto.Int32(100),
				End:   proto.Int32(200),
			}},
		}},
		Extension: []*descriptorpb.FieldDescriptorProto{{
			Name:     proto.String("trace_id"),
			Number:   proto.Int32(100),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			Extendee: proto.String(".loaded.v1.Event"),
		}},
	}}})
	require.NoError(t, err)
	fd, err := files.FindFileByPath("loaded/v1/loaded.proto")
	require.NoError(t, err)
	require.NoError(t, reg.RegisterFile(fd))

	desc, err := reg.FindDescriptorByName("loaded.v1.Event")
	require.NoError(t, err)
	md, ok := desc.(protoreflect.MessageDescriptor)
	require.True(t, ok)

	xt, err := reg.Types().FindExtensionByName("loaded.v1.trace_id")
	require.NoError(t, err, "the registry did not pick up the extension")

	var sawExtension bool
	for range 50 {
		msg := dynamicpb.NewMessage(md)
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{MaxDepth: 5, Extensions: reg.Types()}))
		if !msg.Has(xt.TypeDescriptor()) {
			continue
		}
		sawExtension = true
		assert.NotEmpty(t, msg.Get(xt.TypeDescriptor()).String())

		// The JSON names the extension, so reading it back needs the registry's
		// resolver. Without it protojson calls the key an unknown field.
		jsonBytes, err := protojson.Marshal(msg)
		require.NoError(t, err)
		assert.Contains(t, string(jsonBytes), "[loaded.v1.trace_id]")
		require.Error(t, protojson.Unmarshal(jsonBytes, dynamicpb.NewMessage(md)))
		back := dynamicpb.NewMessage(md)
		require.NoError(t, protojson.UnmarshalOptions{Resolver: reg.Resolver()}.Unmarshal(jsonBytes, back))
		assert.True(t, proto.Equal(msg, back))
	}
	assert.True(t, sawExtension, "the extension was never set")
}
