package main

import (
	"fmt"
	"log"

	ownerv1 "buf.build/gen/go/bufbuild/registry/protocolbuffers/go/buf/registry/owner/v1"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	runtimeapi "buf.build/gen/go/kubernetes/cri-api/protocolbuffers/go/pkg/apis/runtime/v1"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	{
		msg := &elizav1.SayResponse{}
		if err := fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}); err != nil {
			log.Fatalf("err: %s", err)
		}
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}

	{
		msg := &ownerv1.Owner{}
		if err := fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}); err != nil {
			log.Fatalf("err: %s", err)
		}
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}

	{
		msg := &runtimeapi.ListMetricDescriptorsResponse{}
		if err := fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{}); err != nil {
			log.Fatalf("err: %s", err)
		}
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}
}
