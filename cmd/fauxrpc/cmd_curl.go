package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/quic-go/quic-go/http3"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	Stubs               []string `help:"Directories or file paths for JSON files."`
	Method              string   `arg:"" help:"Service or method name." optional:""`
}

func (c *CurlCmd) Run(globals *Globals) error {
	c.Addr = strings.TrimSuffix(c.Addr, "/")

	ctx := context.Background()
	reg, err := registry.NewServiceRegistry()
	if err != nil {
		return fmt.Errorf("failed to create server instance: %w", err)
	}

	stubDB := stubs.NewStubDatabase()

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

	for _, path := range c.Stubs {
		if err := stubs.LoadStubsFromFile(reg, stubDB, path); err != nil {
			return err
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
			methodsToCall[c.Method] = methodDesc
		} else {
			// Service name provided, call all non-streaming methods in that service
			serviceName := c.Method
			serviceDesc := reg.Get(serviceName)
			if serviceDesc == nil {
				return fmt.Errorf("service %s not found", serviceName)
			}
			for i := range serviceDesc.Methods().Len() {
				methodDesc := serviceDesc.Methods().Get(i)
				fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
				methodsToCall[fullMethodName] = methodDesc
			}
		}
	} else {
		// Call all non-streaming methods in all discovered services
		reg.ForEachService(func(serviceDesc protoreflect.ServiceDescriptor) bool {
			for i := range serviceDesc.Methods().Len() {
				methodDesc := serviceDesc.Methods().Get(i)
				fullMethodName := fmt.Sprintf("%s/%s", serviceDesc.FullName(), methodDesc.Name())
				methodsToCall[fullMethodName] = methodDesc
			}
			return true
		})
	}

	if len(methodsToCall) == 0 {
		return fmt.Errorf("no methods found to call")
	}

	for fullMethodName, methodDesc := range methodsToCall {
		if err := c.callRPC(ctx, httpClient, fullMethodName, methodDesc, stubDB); err != nil {
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
	stubDB stubs.StubDatabase,
) error {
	slog.Debug("Calling RPC", "method", fullMethodName)

	stubFaker := stubs.NewStubFaker(stubDB)
	fauxFaker := fauxrpc.NewFauxFaker()
	multiFaker := fauxrpc.NewMultiFaker([]fauxrpc.ProtoFaker{stubFaker, fauxFaker})

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

	client := connect.NewClient[dynamicpb.Message, dynamicpb.Message](
		httpClient,
		c.Addr+"/"+fullMethodName,
		options...,
	)

	isClientStream := methodDesc.IsStreamingClient()
	isServerStream := methodDesc.IsStreamingServer()

	reqMsg := dynamicpb.NewMessage(methodDesc.Input()).New().Interface().(*dynamicpb.Message)
	celCtx := &protocel.CELContext{
		MethodDescriptor: methodDesc,
		Req:              reqMsg,
	}
	genOpts := fauxrpc.GenOptions{
		MaxDepth: 20,
		Faker:    gofakeit.New(0),
		Context:  protocel.WithCELContext(ctx, celCtx),
	}

	// Helper to print request
	printRequest := func(msg proto.Message) {
		jsonBytes, err := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}.Marshal(msg)
		if err != nil {
			slog.Error("Failed to marshal request to JSON", "error", err)
			return
		}
		fmt.Printf("-> [%s]:\n%s\n\n", fullMethodName, string(jsonBytes))
	}

	// Helper to print response
	printResponse := func(msg proto.Message) {
		jsonBytes, err := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}.Marshal(msg)
		if err != nil {
			slog.Error("Failed to marshal response to JSON", "error", err)
			return
		}
		fmt.Printf("<- [%s]:\n%s\n\n", fullMethodName, string(jsonBytes))
	}

	// Helper to print error
	printError := func(err error) {
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
	}

	if !isClientStream && !isServerStream {
		if err := multiFaker.SetDataOnMessage(reqMsg, genOpts); err != nil {
			return err
		}
		printRequest(reqMsg)

		resp, err := client.CallUnary(ctx, connect.NewRequest(reqMsg))
		if err != nil {
			printError(err)
			return nil
		}
		if resp.Msg == nil {
			m := dynamicpb.NewMessage(methodDesc.Output()).New().Interface().(*dynamicpb.Message)
			resp.Msg = m
		}
		slog.Debug("RPC call successful", "method", fullMethodName, "msg", resp.Msg)
		printResponse(resp.Msg)
		return nil
	}

	// Streaming logic
	streamEntry := &stubs.StreamEntry{
		Items: []stubs.StreamItemEntry{{}}, // One empty item to trigger generation
	}

	var fallbackGenerator stubs.FallbackGenerator = func(msg proto.Message) error {
		return multiFaker.SetDataOnMessage(msg, genOpts)
	}

	// For streaming, we need to check if we have a stub that defines a stream
	if isClientStream {
		stubEntry, err := stubFaker.FindStub(ctx, celCtx, methodDesc.Input())
		if err == nil && stubEntry != nil {
			if stubEntry.Stream != nil {
				streamEntry = stubEntry.Stream
				fallbackGenerator = nil // Use explicit stream, no fallback
			} else {
				// Unary stub for stream? Treat as default stream but let multiFaker pick up the stub content
				// streamEntry remains as default
				// fallbackGenerator remains as multiFaker (which includes stubFaker)
			}
		}
	}

	var streamSender func(proto.Message) error
	var streamReceiver func() error
	var bidiStream *connect.BidiStreamForClient[dynamicpb.Message, dynamicpb.Message]
	var clientStream *connect.ClientStreamForClient[dynamicpb.Message, dynamicpb.Message]
	var serverStream *connect.ServerStreamForClient[dynamicpb.Message]

	if isClientStream && isServerStream {
		bidiStream = client.CallBidiStream(ctx)
		streamSender = func(msg proto.Message) error {
			printRequest(msg)
			dm := msg.(*dynamicpb.Message)
			return bidiStream.Send(dm)
		}
		streamReceiver = func() error {
			for {
				msg, err := bidiStream.Receive()
				if err != nil {
					return err
				}
				printResponse(msg)
			}
		}
	} else if isClientStream {
		clientStream = client.CallClientStream(ctx)
		streamSender = func(msg proto.Message) error {
			printRequest(msg)
			dm := msg.(*dynamicpb.Message)
			return clientStream.Send(dm)
		}
		streamReceiver = func() error {
			resp, err := clientStream.CloseAndReceive()
			if err != nil {
				return err
			}
			printResponse(resp.Msg)
			return nil
		}
	} else if isServerStream {
		// Server streaming requires an initial request
		if err := multiFaker.SetDataOnMessage(reqMsg, genOpts); err != nil {
			return err
		}
		printRequest(reqMsg)
		var err error
		serverStream, err = client.CallServerStream(ctx, connect.NewRequest(reqMsg))
		if err != nil {
			printError(err)
			return nil
		}
		streamReceiver = func() error {
			for {
				if !serverStream.Receive() {
					return serverStream.Err()
				}
				printResponse(serverStream.Msg())
			}
		}
	}

	// Execute sending if client streaming involved
	if isClientStream {
		err := stubs.ExecuteStream(ctx, streamEntry, methodDesc.Input(), celCtx, streamSender, fallbackGenerator)
		if err != nil {
			printError(err)
			return nil
		}
		// Close send direction
		if bidiStream != nil {
			if err := bidiStream.CloseRequest(); err != nil {
				printError(err)
				return nil
			}
		}
		// clientStream.CloseAndReceive() is called in receiver
	}

	// Receive responses
	if streamReceiver != nil {
		if err := streamReceiver(); err != nil {
			if !errors.Is(err, context.Canceled) {
				if errors.Is(err, io.EOF) {
					return nil
				}
				printError(err)
			}
		}
	}

	return nil
}
