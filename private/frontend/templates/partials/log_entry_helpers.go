package partials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type HeaderPair struct {
	Key   string
	Value string
}

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

func stubFilename(entry *log.LogEntry) string {
	_, svc := splitService(entry.Service)
	raw := strings.ToLower(svc + "-" + entry.Method)
	slug := strings.Trim(nonAlphaNumRegex.ReplaceAllString(raw, "-"), "-")
	if slug == "" {
		slug = "fauxrpc"
	}
	shortID := entry.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	if shortID != "" {
		return fmt.Sprintf("%s-%s.stub.yaml", slug, strings.ToLower(shortID))
	}
	return fmt.Sprintf("%s.stub.yaml", slug)
}

func splitService(fullService string) (pkg string, svc string) {
	idx := strings.LastIndex(fullService, ".")
	if idx == -1 {
		return "", fullService
	}
	return fullService[:idx], fullService[idx+1:]
}

func formatCompactDuration(d time.Duration) string {
	if d <= 0 {
		return "0µs"
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		ms := float64(d.Microseconds()) / 1000.0
		if ms < 10 {
			return fmt.Sprintf("%.2fms", ms)
		}
		return fmt.Sprintf("%.1fms", ms)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func parseHeaderPairs(raw json.RawMessage) []HeaderPair {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var headers map[string]json.RawMessage
	if err := json.Unmarshal(raw, &headers); err != nil || len(headers) == 0 {
		return nil
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []HeaderPair
	for _, k := range keys {
		vRaw := headers[k]
		var values []string
		if err := json.Unmarshal(vRaw, &values); err == nil {
			for _, v := range values {
				pairs = append(pairs, HeaderPair{Key: k, Value: v})
			}
			continue
		}
		var value string
		if err := json.Unmarshal(vRaw, &value); err == nil {
			pairs = append(pairs, HeaderPair{Key: k, Value: value})
			continue
		}
		pairs = append(pairs, HeaderPair{Key: k, Value: string(vRaw)})
	}
	return pairs
}

func generateStubYAML(entry *log.LogEntry, reg registry.ServiceRegistry) string {
	target := entry.Service + "/" + entry.Method
	stubEntry := stubs.StubFileEntry{
		ID:     entry.ID,
		Target: target,
	}

	var activeIf string
	if reg != nil {
		// The request may name an extension of the message it was sent to.
		jsonOpts := protojson.UnmarshalOptions{Resolver: reg.Resolver()}
		sd := reg.Get(entry.Service)
		if sd != nil {
			method := sd.Methods().ByName(protoreflect.Name(entry.Method))
			if method != nil {
				isClientStream := method.IsStreamingClient()
				if !isClientStream {
					if len(entry.RequestBody) > 0 {
						reqMsg := dynamicpb.NewMessage(method.Input())
						if err := jsonOpts.Unmarshal(entry.RequestBody, reqMsg); err == nil {
							activeIf = stubs.ActiveIfFromProto(reqMsg)
						}
					}
				} else {
					if len(entry.RequestFrames) > 0 {
						reqMsg := dynamicpb.NewMessage(method.Input())
						if err := jsonOpts.Unmarshal(entry.RequestFrames[0], reqMsg); err == nil {
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
