package stubs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/tailscale/hujson"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type StubFile struct {
	Stubs []StubFileEntry `json:"stubs"`
}

func (f StubFile) ToRequest() (*stubsv1.AddStubsRequest, error) {
	stubs := make([]*stubsv1.Stub, len(f.Stubs))
	for i, stub := range f.Stubs {
		if stub.Target == "" {
			return nil, fmt.Errorf(`"target" is required for each stub; missing for stub %d`, i)
		}
		var contentsJSON string
		if stub.Content != nil {
			b, err := json.Marshal(stub.Content)
			if err != nil {
				return nil, err
			}
			contentsJSON = string(b)
		}
		builder := stubsv1.Stub_builder{
			Ref:      stubsv1.StubRef_builder{Id: proto.String(stub.ID), Target: proto.String(stub.Target)}.Build(),
			Priority: proto.Int32(stub.Priority),
		}
		if contentsJSON != "" {
			builder.Json = proto.String(contentsJSON)
		}
		if stub.CelContent != "" {
			builder.CelContent = proto.String(stub.CelContent)
		}
		if stub.ActiveIf != "" {
			builder.ActiveIf = proto.String(stub.ActiveIf)
		}
		if stub.ErrorCode != 0 {
			code := stubsv1.ErrorCode(int32(stub.ErrorCode))
			builder.Error = stubsv1.Error_builder{
				Code:    &code,
				Message: proto.String(stub.ErrorMessage),
			}.Build()
		}
		if stub.Stream != nil {
			streamBuilder := stubsv1.Stream_builder{
				Repeated: proto.Bool(stub.Stream.Repeated),
			}
			if stub.Stream.DoneAfter != "" {
				d, err := time.ParseDuration(stub.Stream.DoneAfter)
				if err != nil {
					return nil, fmt.Errorf("invalid duration %q in stub %d stream done_after: %w", stub.Stream.DoneAfter, i, err)
				}
				streamBuilder.DoneAfter = durationpb.New(d)
			}

			if len(stub.Stream.Items) > 0 {
				streamItems := make([]*stubsv1.StreamItem, len(stub.Stream.Items))
				for j, s := range stub.Stream.Items {
					siBuilder := stubsv1.StreamItem_builder{}
					if s.Delay != "" {
						d, err := time.ParseDuration(s.Delay)
						if err != nil {
							return nil, fmt.Errorf("invalid duration %q in stub %d stream item %d: %w", s.Delay, i, j, err)
						}
						siBuilder.Delay = durationpb.New(d)
					}
					if s.Content != nil {
						b, err := json.Marshal(s.Content)
						if err != nil {
							return nil, err
						}
						siBuilder.Json = proto.String(string(b))
					}
					if s.CelContent != "" {
						siBuilder.CelContent = proto.String(s.CelContent)
					}
					if s.Error != nil {
						code := stubsv1.ErrorCode(int32(s.Error.Code))
						siBuilder.Error = stubsv1.Error_builder{
							Code:    &code,
							Message: proto.String(s.Error.Message),
						}.Build()
					}
					streamItems[j] = siBuilder.Build()
				}
				streamBuilder.Items = streamItems
			}
			builder.Stream = streamBuilder.Build()
		}
		stubs[i] = builder.Build()
	}

	return stubsv1.AddStubsRequest_builder{Stubs: stubs}.Build(), nil
}

type StubFileEntry struct {
	ID           string               `json:"id" yaml:"id,omitempty"`
	Target       string               `json:"target" yaml:"target"`
	Content      any                  `json:"content,omitempty" yaml:"content,omitempty"`
	CelContent   string               `json:"cel_content,omitempty" yaml:"cel_content,omitempty"`
	ActiveIf     string               `json:"active_if,omitempty" yaml:"active_if,omitempty"`
	ErrorCode    int                  `json:"error_code,omitempty" yaml:"error_code,omitempty"`
	ErrorMessage string               `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	Priority     int32                `json:"priority,omitempty" yaml:"priority,omitempty"`
	Stream       *StubFileStreamEntry `json:"stream,omitempty" yaml:"stream,omitempty"`
}

type StubFileStreamEntry struct {
	Items     []StubFileStreamItemEntry `json:"items,omitempty" yaml:"items,omitempty"`
	Repeated  bool                      `json:"repeated,omitempty" yaml:"repeated,omitempty"`
	DoneAfter string                    `json:"done_after,omitempty" yaml:"done_after,omitempty"`
}

type StubFileStreamItemEntry struct {
	Content    any                 `json:"content,omitempty" yaml:"content,omitempty"`
	CelContent string              `json:"cel_content,omitempty" yaml:"cel_content,omitempty"`
	Delay      string              `json:"delay,omitempty" yaml:"delay,omitempty"`
	Error      *StubFileErrorEntry `json:"error,omitempty" yaml:"error,omitempty"`
}

type StubFileErrorEntry struct {
	Code    int    `json:"code" yaml:"code,omitempty"`
	Message string `json:"message" yaml:"message,omitempty"`
}

func LoadStubsFromFile(registry registry.ServiceRegistry, stubdb StubDatabase, stubsPath string) error {
	h := NewHandler(registry, stubdb)
	handleFile := func(path string) error {
		slog.Debug("addStubsFromFile", "path", path)
		stubFile := StubFile{}
		switch filepath.Ext(path) {
		case ".json", ".jsonc":
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			// handle .jsonc format
			if filepath.Ext(path) == ".jsonc" {
				standardContents, err := standardizeJSON(contents)
				if err != nil {
					return fmt.Errorf("standardize.json: %s: %w", path, err)
				}
				contents = standardContents
			}
			decoder := json.NewDecoder(bytes.NewReader(contents))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&stubFile); err != nil {
				return fmt.Errorf("json.Unmarshal: %s: %w", path, err)
			}
		case ".yaml":
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			decoder := yaml.NewDecoder(bytes.NewReader(contents))
			decoder.KnownFields(true)
			if err := decoder.Decode(&stubFile); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}

		req, err := stubFile.ToRequest()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		if _, err := h.AddStubs(context.Background(), connect.NewRequest(req)); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	fi, err := os.Stat(stubsPath)
	if err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return fs.WalkDir(os.DirFS(stubsPath), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return handleFile(filepath.Join(stubsPath, path))
		})
	case mode.IsRegular():
		return handleFile(stubsPath)
	}
	return nil
}

func standardizeJSON(b []byte) ([]byte, error) {
	ast, err := hujson.Parse(b)
	if err != nil {
		return b, err
	}

	ast.Standardize()
	return ast.Pack(), nil

}

var recordLock sync.Mutex

func AppendStubToFile(filePath string, entry StubFileEntry) error {
	recordLock.Lock()
	defer recordLock.Unlock()

	var stubFile StubFile

	// Ensure the parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	isFileExists := false
	if _, err := os.Stat(filePath); err == nil {
		isFileExists = true
	}

	if isFileExists {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read stub file: %w", err)
		}
		if len(bytes.TrimSpace(data)) > 0 {
			if ext == ".yaml" || ext == ".yml" {
				if err := yaml.Unmarshal(data, &stubFile); err != nil {
					return fmt.Errorf("failed to unmarshal yaml: %w", err)
				}
			} else {
				// Handle potential JSONC or standard JSON
				standardized := data
				if ext == ".jsonc" {
					if std, err := standardizeJSON(data); err == nil {
						standardized = std
					}
				}
				if err := json.Unmarshal(standardized, &stubFile); err != nil {
					return fmt.Errorf("failed to unmarshal json: %w", err)
				}
			}
		}
	}

	stubFile.Stubs = append(stubFile.Stubs, entry)

	var out []byte
	var err error
	if ext == ".yaml" || ext == ".yml" {
		out, err = yaml.Marshal(stubFile)
	} else {
		out, err = json.MarshalIndent(stubFile, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal stub file: %w", err)
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		return fmt.Errorf("failed to write stub file: %w", err)
	}

	return nil
}

func RecordSuccessStub(filePath string, target string, activeIf string, respMsg proto.Message) error {
	var content any
	if respMsg != nil {
		b, err := protojson.Marshal(respMsg)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &content); err != nil {
			return err
		}
	}
	entry := StubFileEntry{
		ID:       uuid.New().String(),
		Target:   target,
		ActiveIf: activeIf,
		Content:  content,
		Priority: 10,
	}
	return AppendStubToFile(filePath, entry)
}

func RecordStreamStub(filePath string, target string, activeIf string, items []StubFileStreamItemEntry) error {
	entry := StubFileEntry{
		ID:       uuid.New().String(),
		Target:   target,
		ActiveIf: activeIf,
		Stream: &StubFileStreamEntry{
			Items: items,
		},
		Priority: 10,
	}
	return AppendStubToFile(filePath, entry)
}

func RecordErrorStub(filePath string, target string, activeIf string, code int, message string) error {
	entry := StubFileEntry{
		ID:           uuid.New().String(),
		Target:       target,
		ActiveIf:     activeIf,
		ErrorCode:    code,
		ErrorMessage: message,
		Priority:     10,
	}
	return AppendStubToFile(filePath, entry)
}
