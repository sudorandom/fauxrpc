package stub

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sudorandom/fauxrpc/private/engine"
)

func TestStubMatcher(t *testing.T) {
	reg := NewRegistry()

	reg.AddStub(StubRule{
		Name: "Get User General",
		Target: MatchTarget{
			OperationID: "getUserById",
		},
		Match: StubMatch{
			PathParams: map[string]string{
				"id": "usr_123",
			},
		},
		Response: StubResponse{
			Status: 200,
			Body:   map[string]any{"id": "usr_123", "name": "General User"},
		},
	})

	reg.AddStub(StubRule{
		Name: "Get User Admin Staging",
		Target: MatchTarget{
			OperationID: "getUserById",
		},
		Match: StubMatch{
			PathParams: map[string]string{
				"id": "usr_123",
			},
			QueryParams: map[string]string{
				"role": "admin",
			},
			Headers: map[string]string{
				"X-Test-Env": "staging",
			},
		},
		Response: StubResponse{
			Status: 200,
			Body:   map[string]any{"id": "usr_123", "name": "Admin User"},
		},
	})

	req := &engine.NormalizedRequest{
		OperationID: "getUserById",
		PathParams: map[string]string{
			"id": "usr_123",
		},
		QueryParams: map[string]string{
			"role": "admin",
		},
		Headers: http.Header{
			"X-Test-Env": []string{"staging"},
		},
	}

	matched, found := reg.FindMatch(req)
	assert.True(t, found)
	assert.Equal(t, "Get User Admin Staging", matched.Name)
}
