package stubs

import (
	"context"
	"fmt"
	"strings"

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
	stubs := make([]*stubsv1.Stub, len(req.Msg.Stubs))
	for i, stub := range req.Msg.Stubs {
		ref := stub.GetRef()
		name, err := normalizeTargetName(ref.GetTarget())
		if err != nil {
			return nil, err
		}

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

		ref.Target = string(md.FullName())

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
		stubs[i] = stub
	}

	for i, id := range ids {
		h.db.AddStub(names[i], id, values[i])
	}

	return connect.NewResponse(&stubsv1.AddStubsResponse{Stubs: stubs}), nil
}

// ListStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) ListStubs(ctx context.Context, req *connect.Request[stubsv1.ListStubsRequest]) (*connect.Response[stubsv1.ListStubsResponse], error) {
	ref := req.Msg.GetStubRef()
	targetName, err := normalizeTargetName(ref.GetTarget())
	if err != nil {
		return nil, err
	}
	pbstubs, err := stubsToProto(h.db.ListStubs(targetName, ref.GetId()))
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
		targetName, err := normalizeTargetName(ref.GetTarget())
		if err != nil {
			return nil, err
		}
		h.db.RemoveStub(targetName, ref.GetId())
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
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}

func normalizeTargetName(target string) (protoreflect.FullName, error) {
	switch strings.Count(target, "/") {
	case 0:
	case 1:
		target = strings.ReplaceAll(target, "/", ".")
	default:
		return "", fmt.Errorf("target name has %d slashes when at most one is acceptable", strings.Count(target, "/"))
	}
	return protoreflect.FullName(target), nil
}