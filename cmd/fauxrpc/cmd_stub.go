package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"slices"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	stubsv1 "github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1/stubsv1connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
)

type StubCmd struct {
	Add  StubAddCmd  `cmd:"" help:"Adds a new stub response by method or type"`
	List StubListCmd `cmd:"" help:"List all registered mocks"`
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
		c.ID = gofakeit.LetterN(5)
	}
	stubs = append(stubs, &stubsv1.Stub{
		Ref: &stubsv1.StubRef{
			Id:     c.ID,
			Target: c.Target,
		},
		Content: &stubsv1.Stub_Json{Json: c.JSON},
	},
	)
	_, err := client.AddStubs(context.Background(), connect.NewRequest(&stubsv1.AddStubsRequest{Stubs: stubs}))
	return err
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
		fmt.Printf("%s (%d)\n", name, len(ids))
		slices.Sort(ids)
		for _, id := range ids {
			fmt.Printf(" - %s\n", id)
		}
	}
	return err
}

type StubGetCmd struct {
	Addr string `short:"a" help:"Address to bind to." default:"http://127.0.0.1:6660"`
}

func (c *StubGetCmd) Run(globals *Globals) error {
	client := newStubClient(c.Addr)
	resp, err := client.ListStubs(context.Background(), connect.NewRequest(&stubsv1.ListStubsRequest{}))
	if err != nil {
		return err
	}
	jsonBody, err := protojson.MarshalOptions{Indent: "  "}.Marshal(resp.Msg)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBody))
	return err
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
