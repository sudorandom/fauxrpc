package generator

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
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
