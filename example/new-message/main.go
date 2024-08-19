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
		msg := fauxrpc.NewMessage(elizav1.File_connectrpc_eliza_v1_eliza_proto.Messages().ByName("SayResponse"))
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}

	{
		msg := fauxrpc.NewMessage(ownerv1.File_buf_registry_owner_v1_owner_proto.Messages().ByName("Owner"))
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}

	{
		msg := fauxrpc.NewMessage(runtimeapi.File_pkg_apis_runtime_v1_api_proto.Messages().ByName("ListMetricDescriptorsResponse"))
		b, err := protojson.MarshalOptions{Indent: "  "}.Marshal(msg)
		if err != nil {
			log.Fatalf("err: %s", err)
		}
		fmt.Println(string(b))
	}
}
