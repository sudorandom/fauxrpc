package fauxrpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/fauxrpc"
)

func TestBool(t *testing.T) {
	testCases := []struct {
		name     string
		fd       protoreflect.FieldDescriptor
		expected []bool
	}{
		{
			name:     "always return true",
			fd:       mustCompileField("bool", "bool_val", ""),
			expected: []bool{true},
		},
		{
			name:     "const constraint set to true",
			fd:       mustCompileField("bool", "bool_val", "(buf.validate.field).bool.const = true"),
			expected: []bool{true},
		},
		// TODO: Debug why constraints don't appear to be showing up for the field descriptors generated
		//       in this way.
		// {
		// 	name:     "const constraint set to false",
		// 	fd:       mustCompileField("bool", "bool_val", "(buf.validate.field).bool.const = false"),
		// 	expected: []bool{false},
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fauxrpc.Bool(tc.fd, fauxrpc.GenOptions{})
			assert.Contains(t, tc.expected, result, "unexpected value")
		})
	}
}
