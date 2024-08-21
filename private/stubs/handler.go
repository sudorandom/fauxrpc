package stubs

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	stubsv1 "github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1"
	stubsv1connect "github.com/sudorandom/fauxrpc/private/proto/gen/stubs/v1/stubsv1connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

var _ stubsv1connect.StubsServiceHandler = (*handler)(nil)

type handler struct {
	db       StubDatabase
	registry *registry.ServiceRegistry
}

func NewHandler(db StubDatabase, registry *registry.ServiceRegistry) *handler {
	return &handler{db: db, registry: registry}
}

// AddStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) AddStubs(ctx context.Context, req *connect.Request[stubsv1.AddStubsRequest]) (*connect.Response[stubsv1.AddStubsResponse], error) {
	ids := make([]string, len(req.Msg.Stubs))
	names := make([]protoreflect.FullName, len(req.Msg.Stubs))
	values := make([]protoreflect.ProtoMessage, len(req.Msg.Stubs))
	for i, stub := range req.Msg.Stubs {
		name := protoreflect.FullName(stub.GetRef().GetTarget())
		desc, err := h.registry.Files().FindDescriptorByName(name)
		if err != nil {
			return nil, err
		}
		var md protoreflect.MessageDescriptor
		switch t := desc.(type) {
		case protoreflect.MethodDescriptor:
			name = t.Output().FullName()
			md = t.Output()
		case protoreflect.MessageDescriptor:
			md = t
		case protoreflect.FieldDescriptor:
			return nil, fmt.Errorf("not valid for %T", desc)
		default:
			return nil, fmt.Errorf("not valid for %T", desc)
		}
		fmt.Println(md, md.FullName())
		fmt.Printf("%+T", md)
		msg := newMessage(md).Interface()

		switch t := stub.GetContent().(type) {
		case *stubsv1.Stub_Json:
			if err := protojson.Unmarshal([]byte(t.Json), msg); err != nil {
				return nil, err
			}
		case *stubsv1.Stub_Proto:
			if err := proto.Unmarshal(t.Proto, msg); err != nil {
				return nil, err
			}
		}
		ids[i] = stub.GetRef().GetId()
		names[i] = name
		values[i] = msg
		log.Printf("MSG %T %+v", msg, msg)
	}

	for i, id := range ids {
		h.db.AddStub(names[i], id, values[i])
	}

	return connect.NewResponse(&stubsv1.AddStubsResponse{}), nil
}

// ListStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) ListStubs(ctx context.Context, req *connect.Request[stubsv1.ListStubsRequest]) (*connect.Response[stubsv1.ListStubsResponse], error) {
	ref := req.Msg.GetStubRef()
	pbstubs, err := stubsToProto(h.db.ListStubs(protoreflect.FullName(ref.GetTarget()), ref.GetId()))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&stubsv1.ListStubsResponse{Stubs: pbstubs}), nil
}

// RemoveAllStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) RemoveAllStubs(context.Context, *connect.Request[stubsv1.RemoveAllStubsRequest]) (*connect.Response[stubsv1.RemoveAllStubsResponse], error) {
	h.db.RemoveAllStubs()
	return connect.NewResponse(&stubsv1.RemoveAllStubsResponse{}), nil
}

// RemoveStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) RemoveStubs(ctx context.Context, msg *connect.Request[stubsv1.RemoveStubsRequest]) (*connect.Response[stubsv1.RemoveStubsResponse], error) {
	for _, ref := range msg.Msg.GetStubRefs() {
		h.db.RemoveStub(protoreflect.FullName(ref.GetTarget()), ref.GetId())
	}
	return connect.NewResponse(&stubsv1.RemoveStubsResponse{}), nil
}

func stubsToProto(allStubs map[protoreflect.FullName]map[string]protoreflect.ProtoMessage) ([]*stubsv1.Stub, error) {
	pbStubs := []*stubsv1.Stub{}
	for target, stubs := range allStubs {
		for id, stub := range stubs {
			content, err := protojson.Marshal(stub)
			if err != nil {
				return nil, err
			}
			pbStubs = append(pbStubs, &stubsv1.Stub{
				Ref: &stubsv1.StubRef{
					Id:     id,
					Target: string(target),
				},
				Content: &stubsv1.Stub_Json{Json: string(content)},
			})
		}
	}
	return pbStubs, nil
}

func newMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		log.Printf("SUPPORTED MESSAGE: %s", mt.Descriptor().FullName())
		return true
	})
	fmt.Println("LOOKING FOR", md.FullName())
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		fmt.Println("ERR", err)
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}
