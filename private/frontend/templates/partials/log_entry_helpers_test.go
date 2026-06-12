package partials

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/registry"
)

func mustNewRegistry() registry.ServiceRegistry {
	r, err := registry.NewServiceRegistry()
	if err != nil {
		panic(err)
	}
	_ = r.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto)
	return r
}

func TestGenerateStubYAML(t *testing.T) {
	reg := mustNewRegistry()
	entry := &log.LogEntry{
		ID:             "test-id",
		Timestamp:      time.Now(),
		Service:        "connectrpc.eliza.v1.ElizaService",
		Method:         "Say",
		ClientProtocol: "gRPC",
		Status:         0,
		ResponseBody:   json.RawMessage(`{"sentence": "world"}`),
		RequestBody:    json.RawMessage(`{"sentence": "hello"}`),
	}

	yamlStr := generateStubYAML(entry, reg)
	
	if !strings.Contains(yamlStr, "id: test-id") {
		t.Errorf("Expected id to be in YAML, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "target: connectrpc.eliza.v1.ElizaService/Say") {
		t.Errorf("Expected target to be in YAML, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "sentence: world") {
		t.Errorf("Expected content to be in YAML, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, `active_if: req.sentence == "hello"`) {
		t.Errorf("Expected active_if to be generated, got:\n%s", yamlStr)
	}
	if strings.Contains(yamlStr, "cel_content") {
		t.Errorf("Expected empty fields like cel_content to be omitted, got:\n%s", yamlStr)
	}
}

func TestGenerateStubYAML_Error(t *testing.T) {
	entry := &log.LogEntry{
		ID:             "test-id",
		Timestamp:      time.Now(),
		Service:        "my.Service",
		Method:         "MyMethod",
		ClientProtocol: "gRPC",
		Status:         13, // Internal Error
		ResponseHeaders: json.RawMessage(`{"grpc-message": ["Internal Server Error"]}`),
	}

	yamlStr := generateStubYAML(entry, nil)
	
	if !strings.Contains(yamlStr, "id: test-id") {
		t.Errorf("Expected id to be in YAML, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "error_code: 13") {
		t.Errorf("Expected error_code to be in YAML, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "error_message: Internal Server Error") {
		t.Errorf("Expected error_message to be in YAML, got:\n%s", yamlStr)
	}
	if strings.Contains(yamlStr, "cel_content") {
		t.Errorf("Expected empty fields like cel_content to be omitted, got:\n%s", yamlStr)
	}
}
