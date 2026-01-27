package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
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
		if err := srv.AddFileFromPath(context.Background(), schema); err != nil {
			return err
		}
	}

	if srv.ServiceCount() == 0 && !c.Empty {
		return errors.New("no services found in the given schemas")
	}

	for _, path := range c.Stubs {
		if err := stubs.LoadStubsFromFile(srv, srv, path); err != nil {
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
