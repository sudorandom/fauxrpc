package fauxrpc_test

import (
	"regexp"
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestMapKeyValidation(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
	fd := md.Fields().ByName("string_to_string_map")
	require.NotNil(t, fd)

	pattern := "^[a-z][a-z_]+[a-z]$"
	constraints := &validate.FieldRules{
		Type: &validate.FieldRules_Map{
			Map: &validate.MapRules{
				Keys: &validate.FieldRules{
					Type: &validate.FieldRules_String_{
						String_: &validate.StringRules{
							Pattern: proto.String(pattern),
						},
					},
				},
			},
		},
	}

	fdWithConstraints := createFieldDescriptorWithConstraints(fd, constraints)

	msg := dynamicpb.NewMessage(md)
	opts := fauxrpc.GenOptions{
		MaxDepth: 1,
	}
	val := fauxrpc.Map(msg.ProtoReflect(), fdWithConstraints, opts)
	require.NotNil(t, val)

	m := val.Map()
	re := regexp.MustCompile(pattern)
	// We want at least one item to test validation, but opts.fake().IntRange(0, 4) might return 0.
	// So we might need to run this multiple times or force itemCount.
	// But let's see if it fails first.
	if m.Len() == 0 {
		// Try again if it happened to generate 0 items
		for range 10 {
			val = fauxrpc.Map(msg.ProtoReflect(), fdWithConstraints, opts)
			m = val.Map()
			if m.Len() > 0 {
				break
			}
		}
	}
	require.Greater(t, m.Len(), 0, "Map should have at least one item for testing")

	m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		assert.Regexp(t, re, k.String(), "Key %q does not match pattern %q", k.String(), pattern)
		return true
	})
}

func TestMapAttributesHeuristics(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
	opts := fauxrpc.GenOptions{
		Faker:    gofakeit.New(0),
		MaxDepth: 2,
	}

	for _, name := range []string{"attributes", "attrs", "custom_attr"} {
		fd := md.Fields().ByName(protoreflect.Name(name))
		require.NotNil(t, fd, "field %s should exist", name)

		msg := dynamicpb.NewMessage(md)
		var val *protoreflect.Value
		for range 10 {
			val = fauxrpc.Map(msg.ProtoReflect(), fd, opts)
			if val != nil && val.Map().Len() > 0 {
				break
			}
		}
		require.NotNil(t, val)
		m := val.Map()
		assert.Greater(t, m.Len(), 0)
		m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			keyStr := k.String()
			assert.NotEmpty(t, keyStr)
			assert.Equal(t, strings.ToLower(keyStr), keyStr)
			assert.NotContains(t, keyStr, " ")
			return true
		})
	}
}
