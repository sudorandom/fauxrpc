package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	registryv1 "github.com/sudorandom/fauxrpc/proto/gen/registry/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/registry/v1/registryv1connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type RegistryCmd struct {
	Add       RegistryAddCmd       `cmd:"" help:"Adds new schema to the registry"`
	RemoveAll RegistryRemoveAllCmd `cmd:"" help:"Remove all stubs"`
}

type RegistryAddCmd struct {
	Addr   string   `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
	Schema []string `help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
}

func (c *RegistryAddCmd) Run(globals *Globals) error {
	theRegistry, err := registry.NewServiceRegistry()
	if err != nil {
		return err
	}
	for _, schema := range c.Schema {
		if err := registry.AddServicesFromPath(theRegistry, schema); err != nil {
			if strings.Contains(err.Error(), "name conflict") {
				continue
			}
			return err
		}
	}
	filespb := make([]*descriptorpb.FileDescriptorProto, 0, theRegistry.NumFiles())
	theRegistry.ForEachFile(func(fd protoreflect.FileDescriptor) {
		filespb = append(filespb, protodesc.ToFileDescriptorProto(fd))
	})
	client := newRegistryClient(c.Addr)
	if _, err := client.AddDescriptors(context.Background(), connect.NewRequest(&registryv1.AddDescriptorsRequest{
		Descriptors: &descriptorpb.FileDescriptorSet{
			File: filespb,
		},
	})); err != nil {
		return err
	}

	return nil
}

type RegistryRemoveAllCmd struct {
	Addr string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
}

func (c *RegistryRemoveAllCmd) Run(globals *Globals) error {
	client := newRegistryClient(c.Addr)
	_, err := client.Reset(context.Background(), connect.NewRequest(&registryv1.ResetRequest{}))
	if err != nil {
		return err
	}
	return nil
}

func newRegistryClient(addr string) registryv1connect.RegistryServiceClient {
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	return registryv1connect.NewRegistryServiceClient(client, addr)
}
