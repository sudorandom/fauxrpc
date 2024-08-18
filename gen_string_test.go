package fauxrpc_test

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/stretchr/testify/assert"
	"github.com/sudorandom/fauxrpc"
)

func TestGenerateString(t *testing.T) {
	testCases := []struct {
		name       string
		fd         protoreflect.FieldDescriptor
		expectedFn func(*testing.T, string)
	}{
		{
			name: "return some text",
			fd:   mustCompileField("string", "string_val", ""),
			expectedFn: func(t *testing.T, result string) {
				assert.NotEmpty(t, result)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectedFn(t, fauxrpc.String(tc.fd))
		})
	}
}
