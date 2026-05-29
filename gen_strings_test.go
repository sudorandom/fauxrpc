package fauxrpc_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func TestString(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("ParameterValues")
	stringField := md.Fields().ByName("string_value")
	require.NotNil(t, stringField)

	t.Run("simple string", func(t *testing.T) {
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(stringField, opts)
		assert.NotEmpty(t, s)
	})

	t.Run("const rule", func(t *testing.T) {
		constVal := "fixed_string_value"
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Const: &constVal,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Equal(t, constVal, s)
	})

	t.Run("example rule", func(t *testing.T) {
		examples := []string{"example1", "example2", "example3"}
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Example: examples,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Contains(t, examples, s)
	})

	t.Run("in rule", func(t *testing.T) {
		inValues := []string{"in1", "in2", "in3"}
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					In: inValues,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Contains(t, inValues, s)
	})

	t.Run("len rule", func(t *testing.T) {
		length := uint64(10)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Len: &length,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Len(t, s, int(length))
	})

	t.Run("min_len and max_len rules", func(t *testing.T) {
		minLen := uint64(5)
		maxLen := uint64(15)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					MinLen: &minLen,
					MaxLen: &maxLen,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.GreaterOrEqual(t, len(s), int(minLen))
		assert.LessOrEqual(t, len(s), int(maxLen))
	})

	t.Run("pattern rule", func(t *testing.T) {
		pattern := "^[a-z]{5}$" // 5 lowercase letters
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					Pattern: &pattern,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Regexp(t, regexp.MustCompile(pattern), s)
	})

	t.Run("well_known email", func(t *testing.T) {
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					WellKnown: &validate.StringRules_Email{
						Email: true,
					},
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		// Use a very lenient regex for validation, just checking for @ and non-empty
		assert.Regexp(t, regexp.MustCompile(`^.+@.+$`), s)
	})

	t.Run("well_known uuid", func(t *testing.T) {
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					WellKnown: &validate.StringRules_Uuid{
						Uuid: true,
					},
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.Len(t, s, 36) // UUID format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
		assert.Regexp(t, regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"), s)
	})

	t.Run("field name heuristics - name", func(t *testing.T) {
		// Create a mock field descriptor directly, bypassing createFieldDescriptorWithConstraints
		nameField := &mockFieldDescriptor{
			name: "name",
			kind: protoreflect.StringKind,
		}
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(nameField, opts)
		assert.NotEmpty(t, s) // Should generate a name
	})

	t.Run("field name heuristics - id", func(t *testing.T) {
		// Create a mock field descriptor directly, bypassing createFieldDescriptorWithConstraints
		idField := &mockFieldDescriptor{
			name: "id",
			kind: protoreflect.StringKind,
		}
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(idField, opts)
		assert.Len(t, s, 36) // Should generate a UUID
	})

	t.Run("min_bytes and max_bytes rules", func(t *testing.T) {
		minBytes := uint64(5)
		maxBytes := uint64(15)
		fd := createFieldDescriptorWithConstraints(stringField, &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					MinBytes: &minBytes,
					MaxBytes: &maxBytes,
				},
			},
		})
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		s := fauxrpc.String(fd, opts)
		assert.GreaterOrEqual(t, len([]byte(s)), int(minBytes))
		assert.LessOrEqual(t, len([]byte(s)), int(maxBytes))
	})

	t.Run("tag/tags heuristics", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0), MaxDepth: 2}

		for _, name := range []string{"tag", "custom_tag"} {
			fd := md.Fields().ByName(protoreflect.Name(name))
			require.NotNil(t, fd, "field %s should exist", name)
			s := fauxrpc.String(fd, opts)
			assert.NotEmpty(t, s)
			assert.Equal(t, strings.ToLower(s), s)
			assert.NotContains(t, s, " ")
		}

		tagsFd := md.Fields().ByName("tags")
		require.NotNil(t, tagsFd)
		msg := dynamicpb.NewMessage(md)
		// We want at least one item to test, so we retry if it generates 0 items
		var val *protoreflect.Value
		for range 10 {
			val = fauxrpc.Repeated(msg.ProtoReflect(), tagsFd, opts)
			if val != nil && val.List().Len() > 0 {
				break
			}
		}
		require.NotNil(t, val)
		list := val.List()
		assert.Greater(t, list.Len(), 0)
		for i := 0; i < list.Len(); i++ {
			s := list.Get(i).String()
			assert.NotEmpty(t, s)
			assert.Equal(t, strings.ToLower(s), s)
			assert.NotContains(t, s, " ")
		}
	})

	t.Run("expanded field name heuristics", func(t *testing.T) {
		md := testv1.File_test_v1_test_proto.Messages().ByName("AllTypes")
		require.NotNil(t, md)
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}

		tests := []struct {
			name      string
			fieldName string
			validate  func(t *testing.T, val string)
		}{
			{"email", "email", func(t *testing.T, val string) {
				assert.Contains(t, val, "@")
				assert.Equal(t, strings.ToLower(val), val)
			}},
			{"phone", "phone", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"client_ip", "client_ip", func(t *testing.T, val string) {
				assert.Regexp(t, regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`), val)
			}},
			{"device_mac", "device_mac", func(t *testing.T, val string) {
				assert.Regexp(t, regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`), val)
			}},
			{"user_agent", "user_agent", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"favorite_color", "favorite_color", func(t *testing.T, val string) {
				assert.Regexp(t, regexp.MustCompile(`^#[0-9a-fA-F]{6}$`), val)
			}},
			{"home_address", "home_address", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"city", "city", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"country", "country", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"postal_code", "postal_code", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"profile_bio", "profile_bio", func(t *testing.T, val string) {
				assert.Greater(t, len(strings.Split(val, " ")), 5)
			}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				fd := md.Fields().ByName(protoreflect.Name(tc.fieldName))
				require.NotNil(t, fd, "field %s should exist", tc.fieldName)
				s := fauxrpc.String(fd, opts)
				tc.validate(t, s)
			})
		}
	})

	t.Run("additional heuristics", func(t *testing.T) {
		opts := fauxrpc.GenOptions{Faker: gofakeit.New(0)}
		tests := []struct {
			name     string
			validate func(t *testing.T, val string)
		}{
			{"created_at", func(t *testing.T, val string) {
				_, err := time.Parse(time.RFC3339, val)
				assert.NoError(t, err, "value %q should be RFC3339 timestamp", val)
			}},
			{"birth_date", func(t *testing.T, val string) {
				_, err := time.Parse(time.RFC3339, val)
				assert.NoError(t, err, "value %q should be RFC3339 timestamp", val)
			}},
			{"company", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"job_title", func(t *testing.T, val string) {
				assert.NotEmpty(t, val)
			}},
			{"currency", func(t *testing.T, val string) {
				assert.Len(t, val, 3)
			}},
			{"lang", func(t *testing.T, val string) {
				assert.Len(t, val, 2)
			}},
			{"locale", func(t *testing.T, val string) {
				assert.Contains(t, val, "-")
			}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				fd := &mockFieldDescriptor{
					name: protoreflect.Name(tc.name),
					kind: protoreflect.StringKind,
				}
				s := fauxrpc.String(fd, opts)
				tc.validate(t, s)
			})
		}
	})
}
