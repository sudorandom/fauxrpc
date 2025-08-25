package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"
	"github.com/tailscale/hujson"
	"go.yaml.in/yaml/v3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/proto"
)

type RunCmd struct {
	Schema       []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Addr         string   `short:"a" help:"Address to bind to." default:"127.0.0.1:6660"`
	NoReflection bool     `help:"Disables the server reflection service."`
	NoHTTPLog    bool     `help:"Disables the HTTP log."`
	NoValidate   bool     `help:"Disables protovalidate."`
	NoDocPage    bool     `help:"Disables the documentation page."`
	NoCORS       bool     `help:"Disables CORS headers."`
	HTTPS        bool     `help:"Enables HTTPS, requires cert and certkey"`
	Cert         string   `help:"Path to certificate file"`
	CertKey      string   `help:"Path to certificate key file"`
	HTTP3        bool     `help:"Enables HTTP/3 support."`
	Empty        bool     `help:"Allows the server to run with no services."`
	OnlyStubs    bool     `help:"Only use pre-defined stubs and don't make up fake data."`
	Stubs        []string `help:"Directories or file paths for JSON files."`
	Dashboard    bool     `help:"Enable the admin dashboard."`
}

func (c *RunCmd) Run(globals *Globals) error {
	srv, err := server.NewServer(server.ServerOpts{
		Version:       fullVersion(),
		RenderDocPage: !c.NoDocPage,
		UseReflection: !c.NoReflection,
		WithHTTPLog:   !c.NoHTTPLog,
		WithValidate:  !c.NoValidate,
		OnlyStubs:     c.OnlyStubs,
		Addr:          c.Addr,
		WithDashboard: c.Dashboard,
	})
	if err != nil {
		return err
	}
	for _, schema := range c.Schema {
		if err := srv.AddFileFromPath(schema); err != nil {
			return err
		}
	}

	if srv.ServiceCount() == 0 && !c.Empty {
		return errors.New("no services found in the given schemas")
	}

	for _, path := range c.Stubs {
		if err := addStubsFromFile(srv, path); err != nil {
			return err
		}
	}

	handler, err := srv.Handler()
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	fmt.Printf("FauxRPC (%s) - %d services loaded, %d stubs loaded\n", fullVersion(), srv.ServiceCount(), srv.NumStubs())
	fmt.Printf("Listening on http://%s\n", c.Addr)
	if c.Dashboard {
		fmt.Printf("Dashboard: http://%s/fauxrpc\n", c.Addr)
	}
	if !c.NoDocPage {
		fmt.Printf("OpenAPI documentation: http://%s/fauxrpc/openapi.html\n", c.Addr)
	}
	fmt.Println()
	fmt.Println("Example Commands:")

	eg, _ := errgroup.WithContext(context.Background())
	if c.HTTP3 {
		if !c.NoReflection {
			fmt.Printf("$ buf curl --http3 https://%s --list-methods\n", c.Addr)
		}
		fmt.Printf("$ buf curl --http3 https://%s/[METHOD_NAME]\n", c.Addr)
		if c.Cert == "" || c.CertKey == "" {
			return errors.New("--cert and --cert-key are required if --http3 is set")
		}
		h3srv := http3.Server{
			Addr:    c.Addr,
			Handler: handler,
		}
		eg.Go(func() error {
			return h3srv.ListenAndServeTLS(c.Cert, c.CertKey)
		})
	}
	if c.HTTPS {
		if !c.NoReflection {
			fmt.Printf("$ buf curl https://%s --list-methods\n", c.Addr)
		}
		fmt.Printf("$ buf curl https://%s/[METHOD_NAME]\n", c.Addr)
		if c.Cert == "" || c.CertKey == "" {
			return errors.New("--cert and --cert-key are required if --https is set")
		}
		eg.Go(func() error {
			return server.ListenAndServeTLS(c.Cert, c.CertKey)
		})
	} else {
		if !c.NoReflection {
			fmt.Printf("$ buf curl --http2-prior-knowledge http://%s --list-methods\n", c.Addr)
		}
		fmt.Printf("$ buf curl --http2-prior-knowledge http://%s/[METHOD_NAME]\n", c.Addr)
		eg.Go(server.ListenAndServe)
	}

	slog.Info("Server started.")

	return eg.Wait()
}

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

func addStubsFromFile(srv server.Server, stubsPath string) error {
	h := stubs.NewHandler(srv, srv)
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
			if err := json.Unmarshal(contents, &stubFile); err != nil {
				return fmt.Errorf("json.Unmarshal: %s: %w", path, err)
			}
		case ".yaml":
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if err := yaml.Unmarshal(contents, &stubFile); err != nil {
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
