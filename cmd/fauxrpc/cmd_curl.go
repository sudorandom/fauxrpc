package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/quic-go/quic-go/http3"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/server"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"

	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type CurlCmd struct {
	Schema       []string `help:"RPC schema modules."`
	Addr         string   `short:"a" help:"Address to make a request to" default:"http://127.0.0.1:6660"`
	NoReflection bool     `help:"Disables server reflection."`
	HTTP3        bool     `help:"Enables HTTP/3."`
	Stubs        []string `help:"Stub file paths."`
	Method       string   `arg:"" help:"Service or method name." optional:""`
}

func (c *CurlCmd) Run(globals *Globals) error {
	// 1. Initialize a server instance to load schemas and handle reflection
	// This is a bit of a hack, but it reuses the existing schema loading logic.
	// A dedicated client-side schema loader might be better in the long run.
	srv, err := server.NewServer(server.ServerOpts{
		UseReflection: !c.NoReflection,
	})
	if err != nil {
		return fmt.Errorf("failed to create server instance: %w", err)
	}

	// Load schemas from files
	for _, schemaPath := range c.Schema {
		if err := srv.AddFileFromPath(schemaPath); err != nil {
			return fmt.Errorf("failed to load schema from %s: %w", schemaPath, err)
		}
	}

	// 2. Create a Connect client
	var httpClient *http.Client
	if c.HTTP3 {
		httpClient = &http.Client{
			Transport: &http3.Transport{},
		}
	} else {
		// For plain HTTP, explicitly use h2c (HTTP/2 Cleartext)
		httpClient = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					// If you're also using this client for non-h2c traffic, you may want
					// to delegate to tls.Dial if the network isn't TCP or the addr isn't
					// in an allowlist.
					return net.Dial(network, addr)
				},
				// Don't forget timeouts!
			},
		}
	}

	// Determine the base URL
	baseURL := c.Addr

	// 3. Discover and filter methods
	methodsToCall := make(map[string]protoreflect.MethodDescriptor)

	if c.Method != "" {
		// User provided a specific service or method
		if strings.Contains(c.Method, "/") {
			// Full method name provided
			serviceName := c.Method[:strings.LastIndex(c.Method, "/")]
			methodName := c.Method[strings.LastIndex(c.Method, "/")+1:]

			serviceDesc := srv.Get(serviceName)
			if serviceDesc == nil {
				return fmt.Errorf("service %s not found", serviceName)
			}
			methodDesc := serviceDesc.Methods().ByName(protoreflect.Name(methodName))
			if methodDesc == nil {
				return fmt.Errorf("method %s not found in service %s", methodName, serviceName)
			}
			if !methodDesc.IsStreamingClient() && !methodDesc.IsStreamingServer() {
				methodsToCall[c.Method] = methodDesc
			} else {
				slog.Warn("Skipping streaming method", "method", c.Method)
			}
		} else {
			// Service name provided, call all non-streaming methods in that service
			serviceName := c.Method
			serviceDesc := srv.Get(serviceName)
			if serviceDesc == nil {
				return fmt.Errorf("service %s not found", serviceName)
			}
			for i := 0; i < serviceDesc.Methods().Len(); i++ {
				methodDesc := serviceDesc.Methods().Get(i)
				if !methodDesc.IsStreamingClient() && !methodDesc.IsStreamingServer() {
					fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
					methodsToCall[fullMethodName] = methodDesc
				} else {
					slog.Warn("Skipping streaming method", "method", methodDesc.FullName())
				}
			}
		}
	} else {
		// Call all non-streaming methods in all discovered services
		srv.ForEachService(func(serviceDesc protoreflect.ServiceDescriptor) bool {
			for i := 0; i < serviceDesc.Methods().Len(); i++ {
				methodDesc := serviceDesc.Methods().Get(i)
				if !methodDesc.IsStreamingClient() && !methodDesc.IsStreamingServer() {
					fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
					methodsToCall[fullMethodName] = methodDesc
				} else {
					slog.Warn("Skipping streaming method", "method", methodDesc.FullName())
				}
			}
			return true
		})
	}

	if len(methodsToCall) == 0 {
		return fmt.Errorf("no non-streaming methods found to call")
	}

	fakers := []fauxrpc.ProtoFaker{
		stubs.NewStubFaker(srv.StubDatabase),
		fauxrpc.NewFauxFaker(),
	}
	faker := fauxrpc.NewMultiFaker(fakers)

	ctx := context.Background()

	// 4. Make RPC calls
	for fullMethodName, methodDesc := range methodsToCall {
		slog.Info("Calling RPC", "method", fullMethodName)
		req := dynamicpb.NewMessage(methodDesc.Input()).New().Interface().(*dynamicpb.Message)
		opts := fauxrpc.GenOptions{
			MaxDepth: 20,
			Faker:    gofakeit.New(0),
			Context: protocel.WithCELContext(ctx, &protocel.CELContext{
				MethodDescriptor: methodDesc,
				Req:              req,
			}),
		}
		if err := faker.SetDataOnMessage(req, opts); err != nil {
			return err
		}

		// Create a new client for each method with a custom codec
		// that can handle dynamic responses.
		codec := &dynamicCodec{
			methodDesc: methodDesc,
		}
		client := connect.NewClient[
			*dynamicpb.Message,
			*dynamicpb.Message,
		](
			httpClient,
			baseURL+"/"+fullMethodName,
			connect.WithGRPC(),
			connect.WithCodec(codec),
		)

		// Make the call
		resp, err := client.CallUnary(context.Background(), connect.NewRequest(&req))
		if err != nil {
			slog.Error("RPC call failed", "method", fullMethodName, "error", err)
			continue
		}

		if resp.Msg == nil {
			slog.Error("RPC call failed - empty message", "method", fullMethodName)
			continue
		}

		// Print the response
		slog.Info("RPC call successful", "method", fullMethodName, "msg", resp.Msg)
		jsonBytes, err := protojson.Marshal(*resp.Msg)
		if err != nil {
			slog.Error("Failed to marshal response to JSON", "error", err)
			continue
		}
		fmt.Printf(`Response for %s:\n%s\n\n`, fullMethodName, string(jsonBytes))
	}

	return nil
}

// dynamicCodec implements connect.Codec to handle dynamicpb.Message responses.
type dynamicCodec struct {
	methodDesc protoreflect.MethodDescriptor
}

func (c *dynamicCodec) Name() string {
	return "proto"
}

func (c *dynamicCodec) Marshal(msg any) ([]byte, error) {
	switch m := msg.(type) {
	case proto.Message:
		return proto.Marshal(m)

	case **dynamicpb.Message:
		if *m == nil {
			return nil, fmt.Errorf("cannot marshal nil **dynamicpb.Message")
		}
		return proto.Marshal(*m)

	default:
		return nil, fmt.Errorf("can't marshal %T", msg)
	}
}

// Unmarshal decodes a binary message into a dynamicpb.Message.
func (c *dynamicCodec) Unmarshal(binary []byte, msg any) error {
	// Check if we're unmarshaling into a *dynamicpb.Message.
	// connect-go will pass a pointer to a nil *dynamicpb.Message.
	if ptr, ok := msg.(**dynamicpb.Message); ok {
		// Create a new dynamic message with the correct output descriptor.
		newMsg := dynamicpb.NewMessage(c.methodDesc.Output())
		// Unmarshal into the new message.
		if err := proto.Unmarshal(binary, newMsg); err != nil {
			return err
		}
		// Point the original pointer to the new message.
		*ptr = newMsg
		return nil
	}
	p, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("can't unmarshal into %T", msg)
	}
	return proto.Unmarshal(binary, p)
}
