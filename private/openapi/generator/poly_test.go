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

func TestPolyResolverResolveAllOfSingleScalar(t *testing.T) {
	strSchema := openapi3.NewStringSchema()
	strSchema.Example = "test-value"
	ref := &openapi3.SchemaRef{Value: strSchema}

	result, err := NewPolyResolver().ResolveAllOf(
		NewWalker(true),
		NewGenerationContext(42, 5),
		[]*openapi3.SchemaRef{ref},
	)
	require.NoError(t, err)
	assert.Equal(t, "test-value", result)
}

func TestPolyResolverResolveAllOfMergedObjects(t *testing.T) {
	schema1 := openapi3.NewObjectSchema().WithProperty("propA", openapi3.NewStringSchema())
	schema1.Properties["propA"].Value.Example = "valA"
	schema2 := openapi3.NewObjectSchema().WithProperty("propB", openapi3.NewStringSchema())
	schema2.Properties["propB"].Value.Example = "valB"

	result, err := NewPolyResolver().ResolveAllOf(
		NewWalker(true),
		NewGenerationContext(42, 5),
		[]*openapi3.SchemaRef{{Value: schema1}, {Value: schema2}},
	)
	require.NoError(t, err)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "valA", resultMap["propA"])
	assert.Equal(t, "valB", resultMap["propB"])
}
