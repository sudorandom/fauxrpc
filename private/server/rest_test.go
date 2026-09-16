package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	connectv2 "connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/vanguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/genproto/googleapis/api/annotations"
)

// TestHandler_REST serves a method over REST via vanguard.Mount, using
// WithRules because the eliza schema carries no google.api.http annotations.
func TestHandler_REST(t *testing.T) {
	reg := mustNewRegistry()
	require.NoError(t, reg.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
	srv, err := NewServer(ServerOpts{Addr: "127.0.0.1:0"})
	require.NoError(t, err)
	srv.ServiceRegistry = reg

	sd := reg.Get("connectrpc.eliza.v1.ElizaService")
	require.NotNil(t, sd)

	rpcServer := connectv2.NewServer()
	rpcServer.Register(buildMethods(sd, fauxrpc.NewFauxFaker(), nil, srv, srv.logger, 5)...)

	resolver := reg.Resolver()
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpcServer,
		connecthttp.WithCodecs(newBinaryCodec(resolver), newJSONCodec(resolver)))
	require.NoError(t, vanguard.Mount(mux, rpcServer,
		vanguard.WithTypeResolver(resolver),
		vanguard.WithRules(&annotations.HttpRule{
			Selector: "connectrpc.eliza.v1.ElizaService.Say",
			Pattern:  &annotations.HttpRule_Post{Post: "/v1/say"},
			Body:     "*",
		})))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/v1/say", "application/json", strings.NewReader(`{"sentence": "hello"}`))
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "fake", resp.Header.Get("X-Fauxrpc-Source"))

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	sentence, ok := body["sentence"].(string)
	require.True(t, ok, "response missing sentence field: %v", body)
	assert.NotEmpty(t, sentence)
}
