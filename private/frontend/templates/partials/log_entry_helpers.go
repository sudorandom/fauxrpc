package partials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func generateStubYAML(entry *log.LogEntry, reg registry.ServiceRegistry) string {
	target := entry.Service + "/" + entry.Method
	stubEntry := stubs.StubFileEntry{
		ID:     entry.ID,
		Target: target,
	}

	var activeIf string
	if reg != nil {
		sd := reg.Get(entry.Service)
		if sd != nil {
			method := sd.Methods().ByName(protoreflect.Name(entry.Method))
			if method != nil {
				isClientStream := method.IsStreamingClient()
				if !isClientStream {
					if len(entry.RequestBody) > 0 {
						reqMsg := dynamicpb.NewMessage(method.Input())
						if err := protojson.Unmarshal(entry.RequestBody, reqMsg); err == nil {
							activeIf = stubs.ActiveIfFromProto(reqMsg)
						}
					}
				} else {
					if len(entry.RequestFrames) > 0 {
						reqMsg := dynamicpb.NewMessage(method.Input())
						if err := protojson.Unmarshal(entry.RequestFrames[0], reqMsg); err == nil {
							activeIf = stubs.ActiveIfFromProto(reqMsg)
						}
					}
				}
			}
		}
	}
	stubEntry.ActiveIf = activeIf

	var rawHeaders map[string][]string
	_ = json.Unmarshal(entry.ResponseHeaders, &rawHeaders)

	// Canonicalize keys
	headers := make(http.Header)
	for k, v := range rawHeaders {
		for _, val := range v {
			headers.Add(k, val)
		}
	}

	if entry.Status != 0 {
		errMsg := headers.Get("Grpc-Message")
		if decoded, err := url.QueryUnescape(errMsg); err == nil {
			errMsg = decoded
		}
		if errMsg == "" && len(entry.ResponseBody) > 0 {
			var errObj struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(entry.ResponseBody, &errObj); err == nil {
				errMsg = errObj.Message
			}
		}
		stubEntry.ErrorCode = entry.Status
		stubEntry.ErrorMessage = errMsg
	} else {
		if len(entry.ResponseBody) > 0 {
			var content any
			if err := json.Unmarshal(entry.ResponseBody, &content); err == nil {
				stubEntry.Content = content
			} else {
				stubEntry.Content = string(entry.ResponseBody)
			}
		} else if len(entry.ResponseFrames) > 0 {
			var items []stubs.StubFileStreamItemEntry
			for _, frame := range entry.ResponseFrames {
				var content any
				if err := json.Unmarshal(frame, &content); err == nil {
					items = append(items, stubs.StubFileStreamItemEntry{
						Content: content,
					})
				}
			}
			if len(items) > 0 {
				stubEntry.Stream = &stubs.StubFileStreamEntry{
					Items: items,
				}
			}
		}
	}

	stubFile := stubs.StubFile{
		Stubs: []stubs.StubFileEntry{stubEntry},
	}

	b, err := yaml.Marshal(stubFile)
	if err != nil {
		return fmt.Sprintf("# Error generating YAML: %v", err)
	}
	return string(b)
}
