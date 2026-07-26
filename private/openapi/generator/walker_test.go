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

	walker := NewWalker()
	status, payload, err := walker.GenerateFromOperation("GET", "/user", "getUser", op, 2)
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

	walker := NewWalker()
	status, payload, err := walker.GenerateFromOperation("GET", "/tree", "getTree", op, 2)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.NotNil(t, payload)
}
