package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"
	"github.com/sudorandom/fauxrpc/private/protobuf"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RunCmd struct {
	Schema []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Addr   string   `short:"a" help:"Address to bind to." default:"127.0.0.1:6660"`
}

func (c *RunCmd) Run(globals *Globals) error {
	registry := protobuf.NewServiceRegistry()

	for _, schema := range c.Schema {
		if err := protobuf.AddServicesFromPath(registry, schema); err != nil {
			return err
		}
	}

	if registry.ServiceCount() == 0 {
		return errors.New("no services found in the given schemas")
	}

	// TODO: Add --no-reflection option
	// TODO: Load descriptors from stdin (assume protocol descriptors in binary format)
	// TODO: way more options for data generator, including a stub service for registering stubs

	serviceNames := []string{}
	vgservices := []*vanguard.Service{}
	registry.ForEachService(func(sd protoreflect.ServiceDescriptor) {
		vgservice := vanguard.NewServiceWithSchema(
			sd, protobuf.NewHandler(sd),
			vanguard.WithTargetProtocols(vanguard.ProtocolGRPC),
			vanguard.WithTargetCodecs(vanguard.CodecProto))
		vgservices = append(vgservices, vgservice)
		serviceNames = append(serviceNames, string(sd.FullName()))
	})

	transcoder, err := vanguard.NewTranscoder(vgservices)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	reflector := grpcreflect.NewReflector(&staticNames{names: serviceNames}, grpcreflect.WithDescriptorResolver(registry.Files()))

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	server := &http.Server{
		Addr:    c.Addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	slog.Info(fmt.Sprintf("Listening on http://%s", c.Addr))
	slog.Info(fmt.Sprintf("See available methods: buf curl --http2-prior-knowledge http://%s --list-methods", c.Addr))
	return server.ListenAndServe()
}
