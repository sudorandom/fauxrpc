package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/brianvoe/gofakeit/v7/source"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/grpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GenerateCmd struct {
	Schema []string `required:"" help:"The modules to use for the RPC schema. It can be protobuf descriptors (binpb, json, yaml), a URL for reflection or a directory of descriptors."`
	Target string   `required:"" help:"Protobuf type" example:"'connectrpc.eliza.v1.IntroduceResponse'"`
	Format string   `default:"json" enum:"json,proto,grpc" help:"Format to output"`
	Seed   *uint64  `help:"Seed for random number generator"`
	Stubs  []string `help:"Directories or file paths for JSON files."`
}

func (c *GenerateCmd) Run(globals *Globals) error {
	theRegistry, err := registry.NewServiceRegistry()
	if err != nil {
		return err
	}
	for _, schema := range c.Schema {
		if err := registry.AddServicesFromPath(context.Background(), theRegistry, schema); err != nil {
			if strings.Contains(err.Error(), "name conflict") {
				continue
			}
			return err
		}
	}

	stubDB := stubs.NewStubDatabase()
	for _, path := range c.Stubs {
		if err := stubs.LoadStubsFromFile(theRegistry, stubDB, path); err != nil {
			return err
		}
	}

	desc, err := theRegistry.FindDescriptorByName(protoreflect.FullName(c.Target))
	if err != nil {
		return err
	}
	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return fmt.Errorf("unexpected type: %T", desc)
	}

	seed := uint64(0)
	if c.Seed == nil {
		seed = uint64(time.Now().UnixNano())
	} else {
		seed = *c.Seed
	}
	fakeSrc := source.NewJSF(seed)
	msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{
		Faker:      gofakeit.NewFaker(fakeSrc, true),
		StubFinder: stubs.NewStubFinder(stubDB),
	})
	if err != nil {
		return err
	}

	switch c.Format {
	case "json":
		jsonBytes, err := protojson.Marshal(msg)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(jsonBytes)
	case "proto":
		protoBytes, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(protoBytes)
	case "grpc":
		protoBytes, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		if err := grpc.WriteGRPCMessage(os.Stdout, protoBytes); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected format: %s", c.Format)
	}

	return nil
}
