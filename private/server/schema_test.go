package server

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testOpenAPISchema = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /hello:
    get:
      responses:
        '200':
          description: OK
`

func TestAddFileFromPathBufDirectoryWithOpenAPI(t *testing.T) {
	if _, err := exec.LookPath("buf"); err != nil {
		t.Skip("buf is required for this test")
	}
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "proto"), 0o755))
	for name, contents := range map[string]string{
		"buf.yaml": "version: v2\nmodules:\n  - path: proto\n",
		"proto/service.proto": `syntax = "proto3";
package test;
message Request {}
message Response {}
service TestService { rpc Call(Request) returns (Response); }
`,
		"openapi.yaml":      testOpenAPISchema,
		"base.openapi.yaml": "openapi: 3.0.3\ninfo:\n  title: Base\n  version: 1.0.0\n",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600))
	}

	srv, err := NewServer(ServerOpts{})
	require.NoError(t, err)
	require.NoError(t, srv.AddFileFromPath(context.Background(), dir))
	require.NotNil(t, srv.Get("test.TestService"))
	require.Zero(t, srv.OpenAPIRouterCount())

	// Explicit OpenAPI inputs still work beside a Buf module.
	require.NoError(t, srv.AddFileFromPath(context.Background(), filepath.Join(dir, "openapi.yaml")))
	require.Equal(t, 1, srv.OpenAPIRouterCount())
}

func TestAddFileFromPathDirectoryDoesNotDiscoverOpenAPI(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested"), 0o755))
	for _, name := range []string{"openapi.yaml", "nested/base.openapi.yaml"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(testOpenAPISchema), 0o600))
	}
	srv, err := NewServer(ServerOpts{})
	require.NoError(t, err)
	require.NoError(t, srv.AddFileFromPath(context.Background(), dir))
	require.Zero(t, srv.OpenAPIRouterCount())

	require.NoError(t, srv.AddFileFromPath(context.Background(), filepath.Join(dir, "nested/base.openapi.yaml")))
	require.Equal(t, 1, srv.OpenAPIRouterCount())
}
