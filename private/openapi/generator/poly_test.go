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
