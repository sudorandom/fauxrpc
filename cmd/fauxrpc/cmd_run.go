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
	connectcors "connectrpc.com/cors"
	"connectrpc.com/validate"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"github.com/tailscale/hujson"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/stubs/v1/stubsv1connect"
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
}

func (c *RunCmd) Run(globals *Globals) error {
	srv, err := server.NewServer(server.ServerOpts{
		Version:       version,
		RenderDocPage: !c.NoDocPage,
		UseReflection: !c.NoReflection,
		WithHTTPLog:   !c.NoHTTPLog,
		WithValidate:  !c.NoValidate,
		OnlyStubs:     c.OnlyStubs,
	})
	if err != nil {
		return err
	}
	for _, schema := range c.Schema {
		if err := srv.AddFileFrompath(schema); err != nil {
			return err
		}
	}

	if srv.ServiceCount() == 0 && !c.Empty {
		return errors.New("no services found in the given schemas")
	}

	stubsHandler := stubs.NewHandler(srv, srv)
	for _, path := range c.Stubs {
		if err := addStubsFromFile(stubsHandler, path); err != nil {
			return err
		}
	}

	mux, err := srv.Mux()
	if err != nil {
		return err
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}
	mux.Handle(stubsv1connect.NewStubsServiceHandler(stubsHandler, connect.WithInterceptors(validateInterceptor)))

	mux.Handle(registryv1connect.NewRegistryServiceHandler(registry.NewHandler(srv), connect.WithInterceptors(validateInterceptor)))

	var handler http.Handler = mux
	if !c.NoCORS {
		middleware := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: connectcors.AllowedMethods(),
			AllowedHeaders: connectcors.AllowedHeaders(),
			ExposedHeaders: connectcors.ExposedHeaders(),
		})
		handler = middleware.Handler(handler)
	}

	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	fmt.Printf("FauxRPC (%s) - %d services loaded, %d stubs loaded\n", fullVersion(), srv.ServiceCount(), srv.NumStubs())
	fmt.Printf("Listening on http://%s\n", c.Addr)
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
			Handler: mux,
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
		stubs[i] = &stubsv1.Stub{
			Ref:        &stubsv1.StubRef{Id: stub.ID, Target: stub.Target},
			Content:    &stubsv1.Stub_Json{Json: contentsJSON},
			CelContent: stub.CelContent,
			ActiveIf:   stub.ActiveIf,
			Priority:   stub.Priority,
		}
	}

	return &stubsv1.AddStubsRequest{Stubs: stubs}, nil
}

type StubFileEntry struct {
	ID           string `json:"id"`
	Target       string `json:"target"`
	Content      any    `json:"content,omitempty"`
	CelContent   string `json:"cel_content,omitempty"`
	ActiveIf     string `json:"active_if,omitempty"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Priority     int32  `json:"priority,omitempty"`
}

func addStubsFromFile(h stubsv1connect.StubsServiceHandler, stubsPath string) error {
	addStubFile := func(stubPath string) error {
		slog.Debug("addStubsFromFile", "path", stubPath)
		contents, err := os.ReadFile(stubPath)
		if err != nil {
			return fmt.Errorf("%s: %w", stubPath, err)
		}
		// handle .jsonc format
		if filepath.Ext(stubPath) == ".jsonc" {
			standardContents, err := standardizeJSON(contents)
			if err != nil {
				return fmt.Errorf("standardize.json: %s: %w", stubPath, err)
			}
			contents = standardContents
		}
		stubFile := StubFile{}
		if err := json.Unmarshal(contents, &stubFile); err != nil {
			return fmt.Errorf("json.Unmarshal: %s: %w", stubPath, err)
		}

		req, err := stubFile.ToRequest()
		if err != nil {
			return fmt.Errorf("%s: %w", stubPath, err)
		}

		if _, err := h.AddStubs(context.Background(), connect.NewRequest(req)); err != nil {
			return fmt.Errorf("%s: %w", stubPath, err)
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
			switch filepath.Ext(path) {
			case ".json", ".jsonc":
				return addStubFile(filepath.Join(stubsPath, path))
			}
			return nil
		})
	case mode.IsRegular():
		return addStubFile(stubsPath)
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
