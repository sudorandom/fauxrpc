package generator

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolyResolverDoesNotMutateBranchSchema(t *testing.T) {
	schema := openapi3.NewObjectSchema().WithProperty("name", openapi3.NewStringSchema())
	ref := &openapi3.SchemaRef{Value: schema}

	result, err := NewPolyResolver().evalBranch(NewWalker(true), NewGenerationContext(42, 5), ref)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, schema.AdditionalProperties.Has)
}

func TestPolyResolverSelectsDiscriminatorMapping(t *testing.T) {
	meows := openapi3.NewBoolSchema()
	meows.Example = true
	cat := &openapi3.SchemaRef{
		Ref:   "#/components/schemas/Cat",
		Value: openapi3.NewObjectSchema().WithProperty("meows", meows),
	}
	barks := openapi3.NewBoolSchema()
	barks.Example = true
	dog := &openapi3.SchemaRef{
		Ref:   "#/components/schemas/Dog",
		Value: openapi3.NewObjectSchema().WithProperty("barks", barks),
	}
	discriminator := &openapi3.Discriminator{
		PropertyName: "petType",
		Mapping: map[string]openapi3.MappingRef{
			"cat": {Ref: "#/components/schemas/Cat"},
			"dog": {Ref: "#/components/schemas/Dog"},
		},
	}

	result, err := NewPolyResolver().ResolveOneOfOrAnyOf(
		NewWalker(true),
		NewGenerationContext(42, 5),
		[]*openapi3.SchemaRef{dog, cat},
		discriminator,
	)
	require.NoError(t, err)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cat", resultMap["petType"])
	assert.Equal(t, true, resultMap["meows"])
	assert.NotContains(t, resultMap, "barks")
}
