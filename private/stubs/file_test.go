package stubs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pkgstub "github.com/sudorandom/fauxrpc/private/stub"
)

func TestLoadStubsFromFileRejectsUnknownUnifiedJSONFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stubs.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
  "stubs": [{
    "name": "get pet",
    "target": {"operationId": "getPet"},
    "response": {"status": 200},
    "prioritty": 10
  }]
}`), 0o600))

	unifiedRegistry := pkgstub.NewRegistry()
	database := &unifiedStubDatabase{
		StubDatabase:    NewStubDatabase(),
		unifiedRegistry: unifiedRegistry,
	}

	err := LoadStubsFromFile(nil, database, path)
	require.Error(t, err)
	assert.Zero(t, unifiedRegistry.NumStubs())
}

func TestLoadStubsFromFileDetectsUnifiedTargetsBeyondFirstEntry(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		contents string
	}{
		{
			name: "JSON",
			ext:  ".json",
			contents: `{"stubs":[
  {"name":"gRPC first","target":{"service":"pets.Pets","method":"GetPet"},"response":{"status":200}},
  {"name":"OpenAPI second","target":{"operationId":"getPet"},"response":{"status":200}}
]}`,
		},
		{
			name: "YAML",
			ext:  ".yaml",
			contents: `stubs:
  - name: gRPC first
    target:
      service: pets.Pets
      method: GetPet
    response:
      status: 200
  - name: OpenAPI second
    target:
      operationId: getPet
    response:
      status: 200
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "stubs"+test.ext)
			require.NoError(t, os.WriteFile(path, []byte(test.contents), 0o600))
			unifiedRegistry := pkgstub.NewRegistry()
			database := &unifiedStubDatabase{
				StubDatabase:    NewStubDatabase(),
				unifiedRegistry: unifiedRegistry,
			}

			require.NoError(t, LoadStubsFromFile(nil, database, path))
			assert.Equal(t, 2, unifiedRegistry.NumStubs())
		})
	}
}

type unifiedStubDatabase struct {
	StubDatabase
	unifiedRegistry pkgstub.Registry
}

func (d *unifiedStubDatabase) GetUnifiedRegistry() pkgstub.Registry {
	return d.unifiedRegistry
}
