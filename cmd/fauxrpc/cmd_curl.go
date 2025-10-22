package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/quic-go/quic-go/http3"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/protocel"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type CurlCmd struct {
	Schema              []string `help:"RPC schema modules."`
	Addr                string   `short:"a" help:"Address to make a request to" default:"http://127.0.0.1:6660"`
	Protocol            string   `short:"p" help:"Protocol to use for requests." default:"grpc" enum:"grpc,grpcweb,connect"`
	Encoding            string   `short:"e" help:"Encoding to use for requests." default:"proto" enum:"proto,json"`
	HTTP2PriorKnowledge bool     `name:"http2-prior-knowledge" help:"This flag can be used to indicate that HTTP/2 should be used."`
	HTTP3               bool     `help:"Enables HTTP/3."`
	Method              string   `arg:"" help:"Service or method name." optional:""`
}

func (c *CurlCmd) Run(globals *Globals) error {
	c.Addr = strings.TrimSuffix(c.Addr, "/")

	ctx := context.Background()
	reg, err := registry.NewServiceRegistry()
	if err != nil {
		return fmt.Errorf("failed to create server instance: %w", err)
	}

	var httpClient *http.Client
	if c.HTTP3 {
		httpClient = &http.Client{
			Transport: &http3.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	} else if c.HTTP2PriorKnowledge {
		httpClient = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
	} else {
		httpClient = &http.Client{}
	}

	// Load schemas from files
	for _, schemaPath := range c.Schema {
		if err := registry.AddServicesFromPath(ctx, reg, schemaPath); err != nil {
			return fmt.Errorf("failed to load schema from %s: %w", schemaPath, err)
		}
	}
	if len(c.Schema) == 0 {
		if err := registry.AddServicesFromReflection(reg, httpClient, c.Addr); err != nil {
			return fmt.Errorf("failed to load schema from %s: %w", c.Addr, err)
		}
	}

	methodsToCall := make(map[string]protoreflect.MethodDescriptor)

	if c.Method != "" {
		// User provided a specific service or method
		if strings.Contains(c.Method, "/") {
			// Full method name provided
			serviceName := c.Method[:strings.LastIndex(c.Method, "/")]
			methodName := c.Method[strings.LastIndex(c.Method, "/")+1:]

			serviceDesc := reg.Get(serviceName)
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
				slog.Debug("Skipping streaming method", "method", c.Method)
			}
		} else {
			// Service name provided, call all non-streaming methods in that service
			serviceName := c.Method
			serviceDesc := reg.Get(serviceName)
			if serviceDesc == nil {
				return fmt.Errorf("service %s not found", serviceName)
			}
			for i := range serviceDesc.Methods().Len() {
				methodDesc := serviceDesc.Methods().Get(i)
				if !methodDesc.IsStreamingClient() && !methodDesc.IsStreamingServer() {
					fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
					methodsToCall[fullMethodName] = methodDesc
				} else {
					slog.Debug("Skipping streaming method", "method", methodDesc.FullName())
				}
			}
		}
	} else {
		// Call all non-streaming methods in all discovered services
		reg.ForEachService(func(serviceDesc protoreflect.ServiceDescriptor) bool {
			for i := range serviceDesc.Methods().Len() {
				methodDesc := serviceDesc.Methods().Get(i)
				if !methodDesc.IsStreamingClient() && !methodDesc.IsStreamingServer() {
					fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
					methodsToCall[fullMethodName] = methodDesc
				} else {
					slog.Debug("Skipping streaming method", "method", methodDesc.FullName())
				}
			}
			return true
		})
	}

	if len(methodsToCall) == 0 {
		return fmt.Errorf("no non-streaming methods found to call")
	}

	faker := fauxrpc.NewFauxFaker()

	for fullMethodName, methodDesc := range methodsToCall {
		if err := c.callRPC(ctx, httpClient, fullMethodName, methodDesc, faker); err != nil {
			return err
		}
	}

	return nil
}

func (c *CurlCmd) callRPC(
	ctx context.Context,
	httpClient *http.Client,
	fullMethodName string,
	methodDesc protoreflect.MethodDescriptor,
	faker fauxrpc.ProtoFaker,
) error {
	slog.Debug("Calling RPC", "method", fullMethodName)
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
	requestJsonBytes, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(req)
	if err != nil {
		slog.Error("Failed to marshal response to JSON", "error", err)
		return nil
	}

	fmt.Printf("-> [%s]:\n%s\n\n", fullMethodName, string(requestJsonBytes))

	// Create a new client for each method with a custom codec
	// that can handle dynamic responses.
	options := []connect.ClientOption{}

	switch c.Encoding {
	case "proto":
		options = append(options, connect.WithCodec(&dynamicProtoCodec{
			methodDesc: methodDesc,
		}))
	case "json":
		options = append(options, connect.WithCodec(&dynamicJSONCodec{
			methodDesc: methodDesc,
		}))
	default:
		return fmt.Errorf("unknown encoding: %s", c.Encoding)
	}

	switch c.Protocol {
	case "grpc":
		options = append(options, connect.WithGRPC())
	case "grpcweb":
		options = append(options, connect.WithGRPCWeb())
	case "connect":
		// this is the default
	default:
		return fmt.Errorf("unknown protocol: %s", c.Protocol)
	}

	client := connect.NewClient[*dynamicpb.Message, *dynamicpb.Message](
		httpClient,
		c.Addr+"/"+fullMethodName,
		options...,
	)

	// Make the call
	resp, err := client.CallUnary(ctx, connect.NewRequest(&req))
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			for _, detail := range connectErr.Details() {
				msg, err := detail.Value()
				if err != nil {
					slog.Error("failed to get error detail value", "error", err)
					continue
				}
				jsonBytes, err := protojson.MarshalOptions{
					Multiline: true,
					Indent:    "  ",
				}.Marshal(msg)
				if err != nil {
					slog.Error("Failed to marshal error detail to JSON", "error", err)
					continue
				}
				fmt.Printf("<- [%s] (error) \n%s\n\n", fullMethodName, string(jsonBytes))
			}
		} else {
			fmt.Printf("<- [%s] (error) \n%s\n\n", fullMethodName, err)
		}
		return nil
	}

	if resp.Msg == nil || *resp.Msg == nil {
		m := dynamicpb.NewMessage(methodDesc.Output()).New().Interface().(*dynamicpb.Message)
		resp.Msg = &m
	}

	// Print the response
	slog.Debug("RPC call successful", "method", fullMethodName, "msg", resp.Msg)
	jsonBytes, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(*resp.Msg)
	if err != nil {
		slog.Error("Failed to marshal response to JSON", "error", err)
		return err
	}
	fmt.Printf("<- [%s]:\n%s\n\n", fullMethodName, string(jsonBytes))
	return nil
}
