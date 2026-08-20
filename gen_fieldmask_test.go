package fauxrpc_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func fieldMaskTestField(t *testing.T, msgName, fieldName string) protoreflect.FieldDescriptor {
	t.Helper()
	md := testv1.File_test_v1_test_proto.Messages().ByName(protoreflect.Name(msgName))
	require.NotNil(t, md, "message %s not found", msgName)
	fd := md.Fields().ByName(protoreflect.Name(fieldName))
	require.NotNil(t, fd, "field %s not found in %s", fieldName, msgName)
	return fd
}

func TestGoogleFieldMask(t *testing.T) {
	// A fixed seed keeps the "some path eventually looks like X" assertions from
	// flaking.
	opts := fauxrpc.GenOptions{MaxDepth: 5, Faker: gofakeit.New(1)}

	// target is the message each mask's paths must be valid against.
	tests := map[string]struct {
		msgName   string
		fieldName string
		target    proto.Message
	}{
		"resource named by the request": {
			// UpdateBookRequest.update_mask describes the Book it carries.
			msgName:   "UpdateBookRequest",
			fieldName: "update_mask",
			target:    &testv1.Book{},
		},
		"field named by the mask": {
			// SuffixMaskRequest.book_mask describes book, not the author
			// sitting in front of it.
			msgName:   "SuffixMaskRequest",
			fieldName: "book_mask",
			target:    &testv1.Book{},
		},
		"no resource to describe": {
			// GetBookRequest holds no message field, so the mask describes the
			// request itself.
			msgName:   "GetBookRequest",
			fieldName: "read_mask",
			target:    &testv1.GetBookRequest{},
		},
		"only well-known messages around": {
			msgName:   "ParameterValues",
			fieldName: "field_mask",
			target:    &testv1.ParameterValues{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fd := fieldMaskTestField(t, tt.msgName, tt.fieldName)
			for range 200 {
				mask := fauxrpc.GoogleFieldMask(fd, opts)
				require.NotNil(t, mask)
				assert.NotEmpty(t, mask.GetPaths())
				// IsValid checks every path names a real field of the target,
				// walking only through singular message fields.
				assert.True(t, mask.IsValid(tt.target), "paths %v are not valid for %s", mask.GetPaths(), tt.target.ProtoReflect().Descriptor().FullName())
			}
		})
	}

	t.Run("mask in a collection", func(t *testing.T) {
		// A mask used as a map value hangs off the synthetic map entry, so the
		// request holding the map has to be found before the Book in it can be.
		for range 200 {
			msg := &testv1.MaskCollectionsRequest{}
			require.NoError(t, fauxrpc.SetDataOnMessage(msg, opts))
			for name, mask := range msg.GetMasksByName() {
				assert.NotEmpty(t, mask.GetPaths(), "mask %q is empty", name)
				assert.True(t, mask.IsValid(&testv1.Book{}), "paths %v of mask %q are not valid for test.v1.Book", mask.GetPaths(), name)
			}
			for i, mask := range msg.GetReadMasks() {
				assert.NotEmpty(t, mask.GetPaths(), "mask %d is empty", i)
				assert.True(t, mask.IsValid(&testv1.Book{}), "paths %v of mask %d are not valid for test.v1.Book", mask.GetPaths(), i)
			}
		}
	})

	t.Run("mask generated on its own", func(t *testing.T) {
		// Nothing says what a bare mask masks, so its paths name no real
		// fields. They still have to be paths: protojson rejects anything else.
		for range 200 {
			mask := &fieldmaskpb.FieldMask{}
			require.NoError(t, fauxrpc.SetDataOnMessage(mask, opts))
			assert.NotEmpty(t, mask.GetPaths())
			_, err := protojson.Marshal(mask)
			assert.NoError(t, err, "paths %v", mask.GetPaths())
		}
	})

	t.Run("paths survive JSON", func(t *testing.T) {
		// protojson converts each path to camelCase and refuses any path that
		// does not convert back, which would fail the whole response.
		for _, msgName := range []string{"UpdateBookRequest", "GetBookRequest", "SuffixMaskRequest", "ParameterValues", "AllTypes"} {
			for range 10 {
				msg, err := fauxrpc.NewMessage(testv1.File_test_v1_test_proto.Messages().ByName(protoreflect.Name(msgName)), opts)
				require.NoError(t, err)
				_, err = protojson.Marshal(msg)
				assert.NoError(t, err, "%s does not marshal to JSON", msgName)
			}
		}
	})

	t.Run("does not name itself", func(t *testing.T) {
		fd := fieldMaskTestField(t, "GetBookRequest", "read_mask")
		for range 200 {
			assert.NotContains(t, fauxrpc.GoogleFieldMask(fd, opts).GetPaths(), "read_mask")
		}
	})

	t.Run("paths are normalized", func(t *testing.T) {
		fd := fieldMaskTestField(t, "UpdateBookRequest", "update_mask")
		for range 200 {
			paths := fauxrpc.GoogleFieldMask(fd, opts).GetPaths()
			normalized := &fieldmaskpb.FieldMask{Paths: append([]string{}, paths...)}
			normalized.Normalize()
			// Sorted, deduplicated, and free of paths a broader path covers.
			assert.Equal(t, normalized.GetPaths(), paths)
		}
	})

	t.Run("path shapes", func(t *testing.T) {
		fd := fieldMaskTestField(t, "UpdateBookRequest", "update_mask")
		var sawLeaf, sawNested, sawDeeplyNested bool
		for range 200 {
			for _, path := range fauxrpc.GoogleFieldMask(fd, opts).GetPaths() {
				switch strings.Count(path, ".") {
				case 0:
					sawLeaf = true
				case 1:
					sawNested = true
				default:
					sawDeeplyNested = true
				}
				// Paths stop at repeated fields, maps and well-known types.
				for _, prefix := range []string{"tags.", "labels.", "publish_time."} {
					assert.False(t, strings.HasPrefix(path, prefix), "path %q reaches into %s", path, prefix)
				}
			}
		}
		assert.True(t, sawLeaf, "expected some paths to name a top-level field")
		assert.True(t, sawNested, "expected some paths to reach into a nested message")
		assert.True(t, sawDeeplyNested, "expected some paths to reach two messages down")
	})

	t.Run("depth limit", func(t *testing.T) {
		fd := fieldMaskTestField(t, "UpdateBookRequest", "update_mask")
		for range 200 {
			for _, path := range fauxrpc.GoogleFieldMask(fd, opts).GetPaths() {
				assert.LessOrEqual(t, strings.Count(path, ".")+1, 3, "path %q is too deep", path)
			}
		}
	})

	t.Run("set through FieldValue", func(t *testing.T) {
		fd := fieldMaskTestField(t, "UpdateBookRequest", "update_mask")
		val := fauxrpc.FieldValue(fd, fauxrpc.GenOptions{MaxDepth: 5})
		require.NotNil(t, val)
		require.True(t, val.IsValid())
		mask, ok := val.Message().Interface().(*fieldmaskpb.FieldMask)
		require.True(t, ok)
		assert.True(t, mask.IsValid(&testv1.Book{}), "paths %v are not valid for test.v1.Book", mask.GetPaths())
	})

	t.Run("set through SetDataOnMessage", func(t *testing.T) {
		msg := &testv1.UpdateBookRequest{}
		require.NoError(t, fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{MaxDepth: 5}))
		mask := msg.GetUpdateMask()
		require.NotNil(t, mask)
		assert.NotEmpty(t, mask.GetPaths())
		assert.True(t, mask.IsValid(&testv1.Book{}), "paths %v are not valid for test.v1.Book", mask.GetPaths())
	})
}
