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

	"connectrpc.com/connect"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/tailscale/hujson"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/proto"
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
		stubs[i] = builder.Build()
	}

	return stubsv1.AddStubsRequest_builder{Stubs: stubs}.Build(), nil
}

type StubFileEntry struct {
	ID           string `json:"id" yaml:"id"`
	Target       string `json:"target" yaml:"target"`
	Content      any    `json:"content,omitempty" yaml:"content"`
	CelContent   string `json:"cel_content,omitempty" yaml:"cel_content"`
	ActiveIf     string `json:"active_if,omitempty" yaml:"active_if"`
	ErrorCode    int    `json:"error_code,omitempty" yaml:"error_code"`
	ErrorMessage string `json:"error_message,omitempty" yaml:"error_message"`
	Priority     int32  `json:"priority,omitempty" yaml:"priority"`
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
