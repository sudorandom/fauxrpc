package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/quic-go/quic-go/http3"
	"github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1/stubsv1connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
)

type RunCmd struct {
	Schema       []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Addr         string   `short:"a" help:"Address to bind to." default:"127.0.0.1:6660"`
	NoReflection bool     `help:"Disables the server reflection service."`
	NoDocPage    bool     `help:"Disables the documentation page."`
	HTTPS        bool     `help:"Enables HTTPS, requires cert and certkey"`
	Cert         string   `help:"Path to certificate file"`
	CertKey      string   `help:"Path to certificate key file"`
	HTTP3        bool     `help:"Enables HTTP/3 support."`
}

func (c *RunCmd) Run(globals *Globals) error {
	theRegistry := registry.NewServiceRegistry()

	for _, schema := range c.Schema {
		if err := registry.AddServicesFromPath(theRegistry, schema); err != nil {
			return err
		}
	}

	if theRegistry.ServiceCount() == 0 {
		return errors.New("no services found in the given schemas")
	}
	// TODO: Load descriptors from stdin (assume protocol descriptors in binary format)
	// TODO: add a stub service for registering stubs

	db := stubs.NewStubDatabase()

	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	theRegistry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		vgservice := vanguard.NewServiceWithSchema(
			sd, server.NewHandler(sd, db),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)

	if !c.NoReflection {
		reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(theRegistry.Files()))
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}

	mux.Handle(stubsv1connect.NewStubsServiceHandler(stubs.NewHandler(db, theRegistry), connect.WithInterceptors(validateInterceptor)))

	// OpenAPI Stuff
	if !c.NoDocPage {
		resp, err := convertToOpenAPISpec(theRegistry)
		if err != nil {
			return err
		}
		mux.Handle("GET /fauxrpc/openapi.html", singleFileHandler(openapiHTML))
		mux.Handle("GET /fauxrpc/openapi.yaml", singleFileHandler(resp.File[0].GetContent()))
	}

	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	fmt.Printf("FauxRPC (%s)\n", fullVersion())
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

	return eg.Wait()
}

func singleFileHandler(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}
}
