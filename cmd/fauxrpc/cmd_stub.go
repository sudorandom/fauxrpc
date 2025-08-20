package main

import (
	"cmp"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"

	"connectrpc.com/connect"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/gen/stubs/v1/stubsv1connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type StubCmd struct {
	Add       StubAddCmd       `cmd:"" help:"Adds a new stub response by method or type"`
	List      StubListCmd      `cmd:"" help:"List all registered stubs"`
	Get       StubGetCmd       `cmd:"" help:"Get a registered stub"`
	Remove    StubRemoveCmd    `cmd:"" help:"Remove a registered stub"`
	RemoveAll StubRemoveAllCmd `cmd:"" help:"Remove all stubs"`
}

type StubAddCmd struct {
	Addr         string  `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
	Target       string  `arg:"" help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'"`
	ID           string  `help:"ID to give this particular mock response, will be a random string if one isn't given" example:"bad-response"`
	JSON         string  `help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'"`
	ErrorMessage string  `help:"Message to return with the error"`
	ErrorCode    *uint32 `help:"gRPC Error code to return"`
	ActiveIf     string  `help:"CEL expression that must be true before this mock is used."`
	Priority     int32   `help:"Priority from 0-100 (higher is more preferred)" default:"0"`
	CEL          string  `name:"cel" help:"CEL expression"`
}

func (c *StubAddCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	stub := stubsv1.Stub_builder{
		Ref: stubsv1.StubRef_builder{
			Id:     proto.String(c.ID),
			Target: proto.String(c.Target),
		}.Build(),
		ActiveIf:   proto.String(c.ActiveIf),
		Priority:   proto.Int32(c.Priority),
		CelContent: proto.String(c.CEL),
	}.Build()
	if c.JSON != "" {
		stub.SetJson(c.JSON)
	} else if c.ErrorCode != nil {
		code := stubsv1.ErrorCode(*c.ErrorCode)
		stub.SetError(stubsv1.Error_builder{
			Code:    &code,
			Message: proto.String(c.ErrorMessage),
			Details: []*anypb.Any{},
		}.Build())
	} else {
		return errors.New("one of: --error-code or --json is required")
	}
	resp, err := client.AddStubs(context.Background(), connect.NewRequest(stubsv1.AddStubsRequest_builder{Stubs: []*stubsv1.Stub{stub}}.Build()))
	if err != nil {
		return err
	}
	outputStubs(resp.Msg.GetStubs())
	return nil
}

type StubListCmd struct {
	Addr string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
}

func (c *StubListCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	resp, err := client.ListStubs(context.Background(), connect.NewRequest(&stubsv1.ListStubsRequest{}))
	if err != nil {
		return err
	}
	groupedStubs := map[string][]string{}
	for _, stub := range resp.Msg.GetStubs() {
		ref := stub.GetRef()
		name := ref.GetTarget()
		groupedStubs[name] = append(groupedStubs[name], ref.GetId())
	}

	for name, ids := range groupedStubs {
		fmt.Printf("target=%s\n", name)
		slices.Sort(ids)

		for _, id := range ids {
			fmt.Printf("    id=%s\n", id)
		}
	}
	return err
}

type StubGetCmd struct {
	Addr   string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
	Target string `arg:"" help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'"`
	ID     string `arg:"" help:"ID to give this particular mock response, will be a random string if one isn't given" example:"bad-response"`
}

func (c *StubGetCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	resp, err := client.ListStubs(context.Background(), connect.NewRequest(stubsv1.ListStubsRequest_builder{
		StubRef: stubsv1.StubRef_builder{
			Id:     proto.String(c.ID),
			Target: proto.String(c.Target),
		}.Build(),
	}.Build()))
	if err != nil {
		return err
	}
	outputStubs(resp.Msg.GetStubs())
	return nil
}

type StubRemoveCmd struct {
	Addr   string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
	Target string `arg:"" help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'"`
	ID     string `arg:"" help:"ID to give this particular mock response, will be a random string if one isn't given" example:"bad-response"`
}

func (c *StubRemoveCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	_, err := client.RemoveStubs(context.Background(), connect.NewRequest(stubsv1.RemoveStubsRequest_builder{
		StubRefs: []*stubsv1.StubRef{
			stubsv1.StubRef_builder{
				Id:     proto.String(c.ID),
				Target: proto.String(c.Target),
			}.Build(),
		},
	}.Build()))
	if err != nil {
		return err
	}
	return nil
}

type StubRemoveAllCmd struct {
	Addr string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
}

func (c *StubRemoveAllCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	_, err := client.RemoveAllStubs(context.Background(), connect.NewRequest(&stubsv1.RemoveAllStubsRequest{}))
	if err != nil {
		return err
	}
	return nil
}

type StubForOutput struct {
	Ref          *stubsv1.StubRef `json:"ref,omitempty"`
	Content      any              `json:"content,omitempty"`
	CelContent   any              `json:"cel_content,omitempty"`
	ActiveIf     string           `json:"active_if,omitempty"`
	ErrorCode    int              `json:"error_code,omitempty"`
	ErrorMessage string           `json:"error_message,omitempty"`
	Priority     int32            `json:"priority,omitempty"`
}

func outputStubs(stubs []*stubsv1.Stub) {
	slices.SortFunc(stubs, func(a *stubsv1.Stub, b *stubsv1.Stub) int {
		return cmp.Compare(a.GetRef().GetId(), b.GetRef().GetId())
	})
	for _, stub := range stubs {
		outputStub := StubForOutput{
			Ref:        stub.GetRef(),
			ActiveIf:   stub.GetActiveIf(),
			Priority:   stub.GetPriority(),
			CelContent: stub.GetCelContent(),
		}

		switch stub.WhichContent() {
		case stubsv1.Stub_Json_case:
			var v any
			if err := json.Unmarshal([]byte(stub.GetJson()), &v); err != nil {
				slog.Error("error marshalling for output", slog.Any("error", err))
				continue
			}
			outputStub.Content = v
		case stubsv1.Stub_Error_case:
			outputStub.ErrorCode = int(stub.GetError().GetCode())
			outputStub.ErrorMessage = stub.GetError().GetMessage()
		}
		b, err := json.MarshalIndent(outputStub, "", "  ")
		if err != nil {
			slog.Error("error marshalling for output", slog.Any("error", err))
			continue
		}
		fmt.Println(string(b))
	}
}

func newStubClient(addr string) stubsv1connect.StubsServiceClient {
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	return stubsv1connect.NewStubsServiceClient(client, addr)
}
