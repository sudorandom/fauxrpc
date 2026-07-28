package generator

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkerCycleSafety(t *testing.T) {
	// Self-referencing schema (User has parent User)
	userSchema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"id": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	}
	userSchema.Properties["parent"] = &openapi3.SchemaRef{
		Value: userSchema,
	}

	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Content: openapi3.NewContentWithJSONSchema(userSchema),
				},
			}),
		),
	}

	walker := NewWalker(true)
	status, _, payload, err := walker.GenerateFromOperation("GET", "/user", "getUser", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	resMap, ok := payload.(map[string]any)
	assert.True(t, ok)
	assert.NotNil(t, resMap["id"])
}

func TestWalkerSelectsFallbackResponseDeterministically(t *testing.T) {
	response := func(value string) *openapi3.ResponseRef {
		schema := openapi3.NewStringSchema()
		schema.Example = value
		return &openapi3.ResponseRef{Value: openapi3.NewResponse().WithJSONSchema(schema)}
	}
	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(418, response("teapot")),
			openapi3.WithStatus(404, response("not found")),
			openapi3.WithStatus(202, response("accepted")),
		),
	}

	for range 20 {
		status, _, payload, err := NewWalker(true).GenerateFromOperation("GET", "/jobs", "getJob", op, 2)
		require.NoError(t, err)
		assert.Equal(t, 202, status)
		assert.Equal(t, "accepted", payload)
	}
}

func TestWalkerPreservesSelectedErrorStatus(t *testing.T) {
	schema := openapi3.NewStringSchema()
	schema.Example = "not found"
	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(404, &openapi3.ResponseRef{Value: openapi3.NewResponse().WithJSONSchema(schema)}),
		),
	}

	status, _, payload, err := NewWalker(true).GenerateFromOperation("GET", "/missing", "missing", op, 2)
	require.NoError(t, err)
	assert.Equal(t, 404, status)
	assert.Equal(t, "not found", payload)
}

func TestParseCode(t *testing.T) {
	tests := map[string]int{
		"100":     100,
		"202":     202,
		"404":     404,
		"599":     599,
		"099":     0,
		"600":     0,
		"default": 0,
		"2XX":     0,
		"20":      0,
	}
	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, expected, parseCode(input))
		})
	}
}

func TestWalkerArrayCycleSafety(t *testing.T) {
	nodeSchema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"id": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	}
	nodeSchema.Properties["children"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			Items: &openapi3.SchemaRef{
				Value: nodeSchema,
			},
		},
	}

	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Content: openapi3.NewContentWithJSONSchema(nodeSchema),
				},
			}),
		),
	}

	walker := NewWalker(true)
	status, _, payload, err := walker.GenerateFromOperation("GET", "/tree", "getTree", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.NotNil(t, payload)
}

func TestWalkerArrayLengthUsesSmallCountWithinConstraints(t *testing.T) {
	tests := []struct {
		name     string
		minItems uint64
		maxItems *uint64
		expected int
	}{
		{name: "default", expected: 2},
		{name: "minimum", minItems: 4, expected: 4},
		{name: "minimum safety cap", minItems: 1000, expected: maxGeneratedArrayItems},
		{name: "large maximum", maxItems: uint64Pointer(100), expected: 2},
		{name: "maximum below default", maxItems: uint64Pointer(1), expected: 1},
		{name: "zero maximum", maxItems: uint64Pointer(0), expected: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())
			schema.MinItems = test.minItems
			schema.MaxItems = test.maxItems

			value, err := NewWalker(true).generateSchema(NewGenerationContext(42, 5), schema)
			require.NoError(t, err)
			items, ok := value.([]any)
			require.True(t, ok)
			assert.Len(t, items, test.expected)
		})
	}
}

func uint64Pointer(value uint64) *uint64 {
	return &value
}

func TestWalkerGeneratesResponseHeaders(t *testing.T) {
	rateLimit := openapi3.NewIntegerSchema()
	rateLimit.Example = 250
	expires := openapi3.NewDateTimeSchema()
	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Headers: openapi3.Headers{
						"X-Rate-Limit": {
							Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: &openapi3.SchemaRef{Value: rateLimit}}},
						},
						"X-Expires-After": {
							Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: &openapi3.SchemaRef{Value: expires}}},
						},
					},
				},
			}),
		),
	}

	walker := NewWalker(false)
	nextSeed := int64(0)
	walker.seedSource = func() int64 {
		nextSeed++
		return nextSeed
	}

	status, headers, _, err := walker.GenerateFromOperation("GET", "/login", "login", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "250", headers["X-Rate-Limit"])
	assert.NotEmpty(t, headers["X-Expires-After"])

	_, nextHeaders, _, err := walker.GenerateFromOperation("GET", "/login", "login", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, "250", nextHeaders["X-Rate-Limit"], "explicit examples remain stable")
	assert.NotEqual(t, headers["X-Expires-After"], nextHeaders["X-Expires-After"])

	stableWalker := NewWalker(true)
	_, stableHeaders, _, err := stableWalker.GenerateFromOperation("GET", "/login", "login", op, 2)
	assert.NoError(t, err)
	_, nextStableHeaders, _, err := stableWalker.GenerateFromOperation("GET", "/login", "login", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, stableHeaders["X-Expires-After"], nextStableHeaders["X-Expires-After"])
}
