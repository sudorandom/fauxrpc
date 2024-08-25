package main

import (
	"cmp"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	stubsv1 "github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1/stubsv1connect"
	"golang.org/x/net/http2"
)

type StubCmd struct {
	Add       StubAddCmd       `cmd:"" help:"Adds a new stub response by method or type"`
	List      StubListCmd      `cmd:"" help:"List all registered stubs"`
	Get       StubGetCmd       `cmd:"" help:"Get a registered stub"`
	Remove    StubRemoveCmd    `cmd:"" help:"Remove a registered stub"`
	RemoveAll StubRemoveAllCmd `cmd:"" help:"Remove all stubs"`
}

type StubAddCmd struct {
	Addr   string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
	Target string `arg:"" help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'"`
	ID     string `help:"ID to give this particular mock response, will be a random string if one isn't given" example:"bad-response"`
	JSON   string `help:"Protobuf method or type" example:"'connectrpc.eliza.v1/Say', 'connectrpc.eliza.v1.IntroduceResponse'" required:""`
}

func (c *StubAddCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	stubs := []*stubsv1.Stub{}
	if c.ID == "" {
		c.ID = gofakeit.AdjectiveDescriptive() + "-" + strings.ReplaceAll(gofakeit.Animal(), " ", "-") + gofakeit.DigitN(3)
	}
	stubs = append(stubs, &stubsv1.Stub{
		Ref: &stubsv1.StubRef{
			Id:     c.ID,
			Target: c.Target,
		},
		Content: &stubsv1.Stub_Json{Json: c.JSON},
	},
	)
	resp, err := client.AddStubs(context.Background(), connect.NewRequest(&stubsv1.AddStubsRequest{Stubs: stubs}))
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
	for _, stub := range resp.Msg.Stubs {
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
	resp, err := client.ListStubs(context.Background(), connect.NewRequest(&stubsv1.ListStubsRequest{
		StubRef: &stubsv1.StubRef{
			Id:     c.ID,
			Target: c.Target,
		},
	}))
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
	_, err := client.RemoveStubs(context.Background(), connect.NewRequest(&stubsv1.RemoveStubsRequest{
		StubRefs: []*stubsv1.StubRef{
			{
				Id:     c.ID,
				Target: c.Target,
			},
		},
	}))
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
	Ref     *stubsv1.StubRef `json:"ref,omitempty"`
	Content any              `json:"content,omitempty"`
}

func outputStubs(stubs []*stubsv1.Stub) {
	slices.SortFunc(stubs, func(a *stubsv1.Stub, b *stubsv1.Stub) int {
		return cmp.Compare(a.GetRef().GetId(), b.GetRef().GetId())
	})
	for _, stub := range stubs {
		var v any
		if err := json.Unmarshal([]byte(stub.GetJson()), &v); err != nil {
			slog.Error("error marshalling for output", slog.Any("error", err))
			continue
		}
		outputStub := StubForOutput{
			Ref:     stub.Ref,
			Content: v,
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