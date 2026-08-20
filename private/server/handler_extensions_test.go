package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/proto"
)

// TestHandler_Extensions checks that generated responses carry the extensions
// the loaded schema declares, which takes the registry those extensions were
// registered in reaching generation.
func TestHandler_Extensions(t *testing.T) {
	call := func(t *testing.T) *testv1.Event {
		t.Helper()
		logger := fauxlog.NewLogger()
		reg := mustNewRegistry()
		require.NoError(t, reg.RegisterFile(testv1.File_test_v1_extensions_proto))
		s := &mockServer{
			ServiceRegistry: reg,
			StubDatabase:    stubs.NewStubDatabase(),
			logger:          logger,
		}
		validator, err := protovalidate.New()
		require.NoError(t, err)
		service := testv1.File_test_v1_extensions_proto.Services().ByName("EventService")
		require.NotNil(t, service)
		handler := NewHandler(service, fauxrpc.NewFauxFaker(), validator, s, logger, 20)

		// extensions.proto is an editions file, so its generated messages use
		// the opaque API: fields are set through setters, not struct literals.
		request := &testv1.GetEventRequest{}
		request.SetId("evt_7f3a91")
		var body bytes.Buffer
		writeMsg(t, &body, request)
		req := httptest.NewRequest(http.MethodPost, "/test.v1.EventService/GetEvent", &body)
		req.Header.Set("Content-Type", "application/grpc")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		respMsg := &testv1.GetEventResponse{}
		// Skip the 5 byte gRPC frame prefix.
		require.NoError(t, proto.Unmarshal(response.Body.Bytes()[5:], respMsg))
		require.NotNil(t, respMsg.GetEvent())
		return respMsg.GetEvent()
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
