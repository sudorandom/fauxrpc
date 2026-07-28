package stub

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/engine"
)

func TestRegistryFindMatchReturnsIndependentCopy(t *testing.T) {
	registry := NewRegistry()
	registry.AddStub(StubRule{
		Target: MatchTarget{OperationID: "getPet"},
		Match: StubMatch{Headers: map[string]string{
			"Authorization": "Bearer secret",
		}},
		Response: StubResponse{
			Status:  http.StatusOK,
			Headers: map[string]string{"X-Source": "registry"},
			Body:    map[string]any{"name": "Fluffy"},
		},
	})
	request := &engine.NormalizedRequest{
		OperationID: "getPet",
		Headers:     http.Header{"Authorization": []string{"Bearer secret"}},
	}

	matched, found := registry.FindMatch(request)
	require.True(t, found)
	matched.Target.OperationID = "changed"
	matched.Match.Headers["Authorization"] = "changed"
	matched.Response.Headers["X-Source"] = "changed"
	matched.Response.Body.(map[string]any)["name"] = "changed"

	matchedAgain, found := registry.FindMatch(request)
	require.True(t, found)
	assert.Equal(t, "getPet", matchedAgain.Target.OperationID)
	assert.Equal(t, "Bearer secret", matchedAgain.Match.Headers["Authorization"])
	assert.Equal(t, "registry", matchedAgain.Response.Headers["X-Source"])
	assert.Equal(t, "Fluffy", matchedAgain.Response.Body.(map[string]any)["name"])
}
