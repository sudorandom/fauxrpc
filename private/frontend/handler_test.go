package frontend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/partials"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
)

type mockProvider struct {
	registry.ServiceRegistry
	stats  *metrics.Stats
	logger *log.Logger
	stubDB stubs.StubDatabase
}

func newMockProvider(t *testing.T) *mockProvider {
	t.Helper()
	reg, err := registry.NewServiceRegistry()
	require.NoError(t, err)
	require.NoError(t, reg.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))

	return &mockProvider{
		ServiceRegistry: reg,
		stats: &metrics.Stats{
			TotalRequests:     42,
			RequestsPerSecond: 3,
			Errors:            1,
			ErrorRate:         "2.38",
			UniqueServices:    1,
			UniqueMethods:     3,
			HTTPHost:          "127.0.0.1:6660",
			GoVersion:         "go1.24",
			FauxRpcVersion:    "v0.16.0",
			StartedAt:         time.Now().Add(-5 * time.Minute),
			RequestCounts:     make(map[time.Time]int64),
		},
		logger: log.NewLogger(),
		stubDB: stubs.NewStubDatabase(),
	}
}

func (m *mockProvider) GetStats() *metrics.Stats     { return m.stats }
func (m *mockProvider) GetLogger() *log.Logger       { return m.logger }
func (m *mockProvider) GetStubDB() stubs.StubDatabase { return m.stubDB }
func (m *mockProvider) HasProtobufServices() bool    { return true }
func (m *mockProvider) HasOpenAPIRoutes() bool       { return true }

func TestDashboardRoutesAndRefresh(t *testing.T) {
	p := newMockProvider(t)
	handler := DashboardHandler(p)

	// 1. Primary landing view (/fauxrpc/) defaults to Request Log
	t.Run("PrimaryViewIsRequestLog", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fauxrpc/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, `id="log-entries"`)
		assert.Contains(t, body, `id="log-inspector-pane"`)
		assert.Contains(t, body, `id="summary-stats"`)
		assert.Contains(t, body, `127.0.0.1:6660`)
	})

	// 2. LogEntry renders YAML Stub with Copy, Download, and Register Stub buttons
	t.Run("LogEntryRendersYAMLDownloadAndRegister", func(t *testing.T) {
		entry := &log.LogEntry{
			ID:             "entry-12345678",
			Timestamp:      time.Now(),
			Service:        "connectrpc.eliza.v1.ElizaService",
			Method:         "Say",
			ClientProtocol: "gRPC",
			Status:         0,
			Duration:       1500 * time.Microsecond,
			RequestHeaders: json.RawMessage(`{"content-type":["application/grpc"]}`),
			RequestBody:    json.RawMessage(`{"sentence":"hello eliza"}`),
			ResponseBody:   json.RawMessage(`{"sentence":"hi there"}`),
		}
		var buf bytes.Buffer
		err := partials.LogEntry(entry, p).Render(context.Background(), &buf)
		require.NoError(t, err)

		rendered := buf.String()
		assert.Contains(t, rendered, `downloadStubYAML(`)
		assert.Contains(t, rendered, `elizaservice-say-entry-12.stub.yaml`)
		assert.Contains(t, rendered, `registerStubFromYAML(`)
		assert.Contains(t, rendered, `copyToClipboard('request-body-entry-12345678'`)
		assert.Contains(t, rendered, `copyToClipboard('response-body-entry-12345678'`)
	})

	// 3. Register Stub from YAML endpoint and view in Stubs split layout
	t.Run("RegisterAndInspectAndDeleteStub", func(t *testing.T) {
		yamlPayload := `stubs:
  - id: custom-say-stub
    target: connectrpc.eliza.v1.ElizaService/Say
    active_if: req.sentence == "hello"
    priority: 10
    content:
      sentence: "custom stub response"
`
		req := httptest.NewRequest(http.MethodPost, "/fauxrpc/stubs/register", strings.NewReader(yamlPayload))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "custom-say-stub")
		assert.Equal(t, 1, p.GetStubDB().NumStubs())

		// Fetch Stubs split view
		reqList := httptest.NewRequest(http.MethodGet, "/fauxrpc/stubs/", nil)
		recList := httptest.NewRecorder()
		handler.ServeHTTP(recList, reqList)
		require.Equal(t, http.StatusOK, recList.Code)
		assert.Contains(t, recList.Body.String(), "custom-say-stub")
		assert.Contains(t, recList.Body.String(), "custom stub response")
		assert.Contains(t, recList.Body.String(), "downloadStubYAML")
		assert.Contains(t, recList.Body.String(), "saveStubChanges")

		// Update stub via Structured Form JSON payload
		formJSON := `{"id":"custom-say-stub","target":"connectrpc.eliza.v1.ElizaService.Say","priority":90,"active_if":"req.sentence == \"hi\"","json_content":"{\"sentence\":\"updated from form\"}"}`
		reqUpdateForm := httptest.NewRequest(http.MethodPut, "/fauxrpc/stubs/connectrpc.eliza.v1.ElizaService.Say/custom-say-stub", strings.NewReader(formJSON))
		reqUpdateForm.Header.Set("Content-Type", "application/json")
		recUpdateForm := httptest.NewRecorder()
		handler.ServeHTTP(recUpdateForm, reqUpdateForm)
		require.Equal(t, http.StatusOK, recUpdateForm.Code, recUpdateForm.Body.String())
		assert.Contains(t, recUpdateForm.Body.String(), "updated from form")
		assert.Contains(t, recUpdateForm.Body.String(), "Priority: 90")
		assert.Equal(t, 1, p.GetStubDB().NumStubs())

		// Update stub via YAML payload
		updatedYAML := `stubs:
  - id: custom-say-stub
    target: connectrpc.eliza.v1.ElizaService/Say
    priority: 95
    content:
      sentence: updated from yaml
`
		reqUpdateYAML := httptest.NewRequest(http.MethodPut, "/fauxrpc/stubs/connectrpc.eliza.v1.ElizaService.Say/custom-say-stub", strings.NewReader(updatedYAML))
		reqUpdateYAML.Header.Set("Content-Type", "application/x-yaml")
		recUpdateYAML := httptest.NewRecorder()
		handler.ServeHTTP(recUpdateYAML, reqUpdateYAML)
		require.Equal(t, http.StatusOK, recUpdateYAML.Code, recUpdateYAML.Body.String())
		assert.Contains(t, recUpdateYAML.Body.String(), "updated from yaml")
		assert.Contains(t, recUpdateYAML.Body.String(), "Priority: 95")
		assert.Equal(t, 1, p.GetStubDB().NumStubs())

		// Register a second stub on the same method and verify only one row has selected-row
		secondYAML := `stubs:
  - id: second-say-stub
    target: connectrpc.eliza.v1.ElizaService/Say
    priority: 50
    content:
      sentence: second stub response
`
		reqSecond := httptest.NewRequest(http.MethodPost, "/fauxrpc/stubs/register", strings.NewReader(secondYAML))
		recSecond := httptest.NewRecorder()
		handler.ServeHTTP(recSecond, reqSecond)
		require.Equal(t, http.StatusOK, recSecond.Code)
		assert.Equal(t, 2, p.GetStubDB().NumStubs())

		reqListMulti := httptest.NewRequest(http.MethodGet, "/fauxrpc/stubs/connectrpc.eliza.v1.ElizaService/Say/second-say-stub", nil)
		reqListMulti.Header.Set("HX-Request", "true")
		recListMulti := httptest.NewRecorder()
		handler.ServeHTTP(recListMulti, reqListMulti)
		require.Equal(t, http.StatusOK, recListMulti.Code)
		assert.Equal(t, 1, strings.Count(recListMulti.Body.String(), "stub-entry-row border-b border-slate-800/80 hover:bg-slate-800/50 transition-colors cursor-pointer text-xs focus:outline-none selected-row"))

		// Delete stubs
		reqDel := httptest.NewRequest(http.MethodDelete, "/fauxrpc/stubs/", nil)
		recDel := httptest.NewRecorder()
		handler.ServeHTTP(recDel, reqDel)
		require.Equal(t, http.StatusOK, recDel.Code)
		assert.Equal(t, 0, p.GetStubDB().NumStubs())
		assert.Contains(t, recDel.Body.String(), "stub-new-root")

		// Create new stub directly from Stubs view via POST /fauxrpc/stubs/
		createFormJSON := `{"id":"created-from-list","target":"connectrpc.eliza.v1.ElizaService/Say","priority":25,"active_if":"req.sentence == \"new\"","json_content":"{\"sentence\":\"created in stub list\"}"}`
		reqCreate := httptest.NewRequest(http.MethodPost, "/fauxrpc/stubs/", strings.NewReader(createFormJSON))
		reqCreate.Header.Set("Content-Type", "application/json")
		recCreate := httptest.NewRecorder()
		handler.ServeHTTP(recCreate, reqCreate)
		require.Equal(t, http.StatusOK, recCreate.Code, recCreate.Body.String())
		assert.Equal(t, 1, p.GetStubDB().NumStubs())
		assert.Contains(t, recCreate.Body.String(), "created-from-list")
		assert.Contains(t, recCreate.Body.String(), "created in stub list")

		// Clean up
		reqDelAll := httptest.NewRequest(http.MethodDelete, "/fauxrpc/stubs/", nil)
		recDelAll := httptest.NewRecorder()
		handler.ServeHTTP(recDelAll, reqDelAll)
		require.Equal(t, http.StatusOK, recDelAll.Code)
		assert.Equal(t, 0, p.GetStubDB().NumStubs())
	})

	// 3b. Live Stub Validation Endpoint (/fauxrpc/stubs/validate)
	t.Run("ValidateStubCELAndJSON", func(t *testing.T) {
		// Valid JSON + CEL + active_if
		validPayload := `{"target":"connectrpc.eliza.v1.ElizaService/Say","active_if":"req.sentence == \"hello\"","json_content":"{\"sentence\":\"hi\"}","cel_content":"{'sentence': req.sentence}"}`
		reqValid := httptest.NewRequest(http.MethodPost, "/fauxrpc/stubs/validate", strings.NewReader(validPayload))
		reqValid.Header.Set("Content-Type", "application/json")
		recValid := httptest.NewRecorder()
		handler.ServeHTTP(recValid, reqValid)
		require.Equal(t, http.StatusOK, recValid.Code)
		var validRes stubValidationResponse
		require.NoError(t, json.Unmarshal(recValid.Body.Bytes(), &validRes))
		assert.True(t, validRes.Valid, recValid.Body.String())

		// Invalid JSON field & invalid CEL syntax
		invalidPayload := `{"target":"connectrpc.eliza.v1.ElizaService/Say","active_if":"req.nonexistent == 123","json_content":"{\"unknown_field\":1}","cel_content":"{'sentence': invalid_syntax("}`
		reqInvalid := httptest.NewRequest(http.MethodPost, "/fauxrpc/stubs/validate", strings.NewReader(invalidPayload))
		reqInvalid.Header.Set("Content-Type", "application/json")
		recInvalid := httptest.NewRecorder()
		handler.ServeHTTP(recInvalid, reqInvalid)
		require.Equal(t, http.StatusOK, recInvalid.Code)
		var invalidRes stubValidationResponse
		require.NoError(t, json.Unmarshal(recInvalid.Body.Bytes(), &invalidRes))
		assert.False(t, invalidRes.Valid)
		assert.NotEmpty(t, invalidRes.ActiveIfError)
		assert.NotEmpty(t, invalidRes.JSONError)
		assert.NotEmpty(t, invalidRes.CELError)
	})

	// 4. Schema Browser Master-Detail Split View
	t.Run("SchemaBrowserSplitView", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fauxrpc/browser/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "connectrpc.eliza.v1")
		assert.Contains(t, body, "eliza.proto")
		assert.Contains(t, body, "service ElizaService")
	})

	// 5. Reset Stats and Clear Logs
	t.Run("ResetStatsAndClearLogs", func(t *testing.T) {
		p.GetLogger().Log(&log.LogEntry{ID: "log-1", Timestamp: time.Now()})
		reqClear := httptest.NewRequest(http.MethodPost, "/fauxrpc/logs/clear", nil)
		recClear := httptest.NewRecorder()
		handler.ServeHTTP(recClear, reqClear)
		require.Equal(t, http.StatusNoContent, recClear.Code)
		hist, _, unsub := p.GetLogger().SubscribeWithHistory()
		unsub()
		assert.Empty(t, hist)

		reqReset := httptest.NewRequest(http.MethodPost, "/fauxrpc/stats/reset", nil)
		recReset := httptest.NewRecorder()
		handler.ServeHTTP(recReset, reqReset)
		require.Equal(t, http.StatusOK, recReset.Code)
		assert.Equal(t, int64(0), p.GetStats().TotalRequests)
	})
}
