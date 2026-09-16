package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
	"github.com/sudorandom/fauxrpc/private/gen/test/v1/testv1connect"
	"google.golang.org/protobuf/proto"
)

// TestHandler_Extensions checks that generated responses carry the extensions
// the loaded schema declares, which takes the registry those extensions were
// registered in reaching generation and the response codec.
func TestHandler_Extensions(t *testing.T) {
	_, ts, httpClient := setupTestServer(t, ServerOpts{}, testv1.File_test_v1_extensions_proto)
	client := testv1connect.NewEventServiceClient(httpClient, ts.URL, connect.WithGRPC())

	call := func(t *testing.T) *testv1.Event {
		t.Helper()
		// extensions.proto is an editions file, so its generated messages use
		// the opaque API: fields are set through setters, not struct literals.
		request := &testv1.GetEventRequest{}
		request.SetId("evt_7f3a91")
		resp, err := client.GetEvent(context.Background(), connect.NewRequest(request))
		require.NoError(t, err)
		require.NotNil(t, resp.Msg.GetEvent())
		return resp.Msg.GetEvent()
	}

	var sawExtension bool
	for range 20 {
		event := call(t)
		if proto.HasExtension(event, testv1.E_TraceId) || proto.HasExtension(event, testv1.E_Sampled) {
			sawExtension = true
			break
		}
	}
	assert.True(t, sawExtension, "no response carried an extension")
}
