package openapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOpenAPISpec(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. OpenAPI YAML file
	openapiYaml := filepath.Join(tmpDir, "api.yaml")
	err := os.WriteFile(openapiYaml, []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
`), 0644)
	assert.NoError(t, err)

	assert.True(t, IsOpenAPISpec(openapiYaml))

	// 2. Swagger JSON file
	swaggerJson := filepath.Join(tmpDir, "swagger.json")
	err = os.WriteFile(swaggerJson, []byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Legacy API",
    "version": "1.0.0"
  }
}`), 0644)
	assert.NoError(t, err)

	assert.True(t, IsOpenAPISpec(swaggerJson))

	// 3. Protobuf / Non-OpenAPI file
	nonOpenapi := filepath.Join(tmpDir, "service.json")
	err = os.WriteFile(nonOpenapi, []byte(`{
  "name": "something_else"
}`), 0644)
	assert.NoError(t, err)

	assert.False(t, IsOpenAPISpec(nonOpenapi))
}
