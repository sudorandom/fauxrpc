package registry_test

import (
	"os"
	"path/filepath"
	"testing"

	"buf.build/go/protoyaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestLoadersPreserveGoogleAPIHTTP(t *testing.T) {
	t.Run(".proto file with google.api.http option", func(t *testing.T) {
		tmpDir := t.TempDir()
		protoPath := filepath.Join(tmpDir, "library.proto")
		protoContent := `syntax = "proto3";
package test.library.v1;

import "google/api/annotations.proto";

service LibraryService {
  rpc GetBook(GetBookRequest) returns (Book) {
    option (google.api.http) = {
      get: "/v1/{name=shelves/*/books/*}"
      additional_bindings {
        get: "/v1/books/{name=*}"
      }
    };
  }
}

message GetBookRequest {
  string name = 1;
}

message Book {
  string name = 1;
  string title = 2;
}
`
		require.NoError(t, os.WriteFile(protoPath, []byte(protoContent), 0o600))

		reg, err := registry.NewServiceRegistry()
		require.NoError(t, err)
		require.NoError(t, registry.AddServicesFromSingleFile(reg, protoPath))

		svc := reg.Get("test.library.v1.LibraryService")
		require.NotNil(t, svc)
		method := svc.Methods().ByName("GetBook")
		require.NotNil(t, method)

		opts, ok := method.Options().(*descriptorpb.MethodOptions)
		require.True(t, ok)
		require.True(t, proto.HasExtension(opts, annotations.E_Http))
		rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
		require.True(t, ok)
		assert.Equal(t, "/v1/{name=shelves/*/books/*}", rule.GetGet())
		require.Len(t, rule.GetAdditionalBindings(), 1)
		assert.Equal(t, "/v1/books/{name=*}", rule.GetAdditionalBindings()[0].GetGet())
	})

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Syntax:  proto.String("proto3"),
				Name:    proto.String("test/library/v1/library.proto"),
				Package: proto.String("test.library.v1"),
				Dependency: []string{
					"google/api/annotations.proto",
				},
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("GetBookRequest"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:     proto.String("name"),
								Number:   proto.Int32(1),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
								JsonName: proto.String("name"),
							},
						},
					},
				},
				Service: []*descriptorpb.ServiceDescriptorProto{
					{
						Name: proto.String("LibraryService"),
						Method: []*descriptorpb.MethodDescriptorProto{
							{
								Name:       proto.String("GetBook"),
								InputType:  proto.String(".test.library.v1.GetBookRequest"),
								OutputType: proto.String(".test.library.v1.GetBookRequest"),
								Options:    &descriptorpb.MethodOptions{},
							},
						},
					},
				},
			},
		},
	}
	proto.SetExtension(fds.File[0].Service[0].Method[0].Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{Get: "/v1/{name=shelves/*/books/*}"},
	})

	formats := []struct {
		ext     string
		marshal func(*descriptorpb.FileDescriptorSet) ([]byte, error)
	}{
		{
			ext: ".binpb",
			marshal: func(d *descriptorpb.FileDescriptorSet) ([]byte, error) {
				return proto.Marshal(d)
			},
		},
		{
			ext: ".json",
			marshal: func(d *descriptorpb.FileDescriptorSet) ([]byte, error) {
				return protojson.Marshal(d)
			},
		},
		{
			ext: ".txtpb",
			marshal: func(d *descriptorpb.FileDescriptorSet) ([]byte, error) {
				return prototext.Marshal(d)
			},
		},
		{
			ext: ".yaml",
			marshal: func(d *descriptorpb.FileDescriptorSet) ([]byte, error) {
				return protoyaml.Marshal(d)
			},
		},
	}

	for _, tc := range formats {
		t.Run(tc.ext, func(t *testing.T) {
			data, err := tc.marshal(fds)
			require.NoError(t, err)

			filePath := filepath.Join(t.TempDir(), "descriptors"+tc.ext)
			require.NoError(t, os.WriteFile(filePath, data, 0o600))

			reg, err := registry.NewServiceRegistry()
			require.NoError(t, err)
			require.NoError(t, registry.AddServicesFromSingleFile(reg, filePath))

			svc := reg.Get("test.library.v1.LibraryService")
			require.NotNil(t, svc)
			method := svc.Methods().ByName("GetBook")
			require.NotNil(t, method)

			opts, ok := method.Options().(*descriptorpb.MethodOptions)
			require.True(t, ok)
			require.True(t, proto.HasExtension(opts, annotations.E_Http))
			rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
			require.True(t, ok)
			assert.Equal(t, "/v1/{name=shelves/*/books/*}", rule.GetGet())
		})
	}
}
