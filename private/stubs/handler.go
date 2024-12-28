package stubs

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc/private/registry"
	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
	stubsv1connect "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1/stubsv1connect"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ stubsv1connect.StubsServiceHandler = (*handler)(nil)

type handler struct {
	registry registry.ServiceRegistry
	stubdb   StubDatabase
}

func NewHandler(registry registry.ServiceRegistry, stubdb StubDatabase) *handler {
	return &handler{
		registry: registry,
		stubdb:   stubdb,
	}
}

// AddStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) AddStubs(ctx context.Context, req *connect.Request[stubsv1.AddStubsRequest]) (*connect.Response[stubsv1.AddStubsResponse], error) {
	entries := make([]StubEntry, len(req.Msg.GetStubs()))
	stubs := make([]*stubsv1.Stub, len(req.Msg.GetStubs()))
	for i, stub := range req.Msg.GetStubs() {
		if !stub.HasRef() {
			stub.SetRef(&stubsv1.StubRef{})
		}
		if stub.GetRef().GetId() == "" {
			stub.GetRef().SetId(gofakeit.AdjectiveDescriptive() + "-" + strings.ReplaceAll(gofakeit.Animal(), " ", "-") + gofakeit.DigitN(3))
		}

		ref := stub.GetRef()
		name, err := normalizeTargetName(ref.GetTarget())
		if err != nil {
			return nil, err
		}

		entry := StubEntry{
			Priority: int(stub.GetPriority()),
		}

		desc, err := h.registry.FindDescriptorByName(name)
		if err != nil {
			return nil, fmt.Errorf("unable to find object named %s: %w", name, err)
		}
		var md protoreflect.MessageDescriptor
		switch t := desc.(type) {
		case protoreflect.MethodDescriptor:
			if len(stub.GetActiveIf()) > 0 {
				r, err := NewActiveIf(t, stub.GetActiveIf())
				if err != nil {
					return nil, err
				}
				entry.ActiveIf = r
			}

			// name = t.Output().FullName()
			md = t.Output()
		case protoreflect.MessageDescriptor:
			md = t
		case protoreflect.FieldDescriptor:
			return nil, fmt.Errorf("not valid for %T", desc)
		default:
			return nil, fmt.Errorf("not valid for %T", desc)
		}

		ref.SetTarget(string(md.FullName()))
		entry.Key = StubKey{ID: stub.GetRef().GetId(), Name: name}

		switch stub.WhichContent() {
		case stubsv1.Stub_Json_case:
			if stub.GetJson() != "" {
				msg := registry.NewMessage(md).Interface()
				if err := protojson.Unmarshal([]byte(stub.GetJson()), msg); err != nil {
					return nil, err
				}
				entry.Message = msg
			}
		case stubsv1.Stub_Proto_case:
			msg := registry.NewMessage(md).Interface()
			if err := proto.Unmarshal(stub.GetProto(), msg); err != nil {
				return nil, err
			}
			entry.Message = msg
		case stubsv1.Stub_Error_case:
			entry.Error = &StatusError{StubsError: stub.GetError()}
		}

		if stub.GetCelContent() != "" {
			celmsg, err := protocel.New(h.registry.Files(), md, stub.GetCelContent())
			if err != nil {
				return nil, err
			}
			entry.CELMessage = celmsg
		}

		entries[i] = entry
		stubs[i] = stub
	}

	for _, entry := range entries {
		h.stubdb.AddStub(entry)
	}

	return connect.NewResponse(stubsv1.AddStubsResponse_builder{Stubs: stubs}.Build()), nil
}

// ListStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) ListStubs(ctx context.Context, req *connect.Request[stubsv1.ListStubsRequest]) (*connect.Response[stubsv1.ListStubsResponse], error) {
	ref := req.Msg.GetStubRef()
	targetName, err := normalizeTargetName(ref.GetTarget())
	if err != nil {
		return nil, err
	}
	filtered := []StubEntry{}
	for _, stub := range h.stubdb.GetStubs() {
		if ref.GetTarget() != "" && targetName != stub.Key.Name {
			continue
		}
		if ref.GetId() != "" && ref.GetId() != stub.Key.ID {
			continue
		}
		filtered = append(filtered, stub)
	}

	pbstubs, err := stubsToProto(filtered)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(stubsv1.ListStubsResponse_builder{Stubs: pbstubs}.Build()), nil

}

// RemoveAllStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) RemoveAllStubs(context.Context, *connect.Request[stubsv1.RemoveAllStubsRequest]) (*connect.Response[stubsv1.RemoveAllStubsResponse], error) {
	h.stubdb.RemoveAllStubs()
	return connect.NewResponse(&stubsv1.RemoveAllStubsResponse{}), nil
}

// RemoveStubs implements stubsv1connect.StubsServiceHandler.
func (h *handler) RemoveStubs(ctx context.Context, msg *connect.Request[stubsv1.RemoveStubsRequest]) (*connect.Response[stubsv1.RemoveStubsResponse], error) {
	for _, ref := range msg.Msg.GetStubRefs() {
		targetName, err := normalizeTargetName(ref.GetTarget())
		if err != nil {
			return nil, err
		}
		h.stubdb.RemoveStub(StubKey{
			Name: targetName,
			ID:   ref.GetId(),
		})
	}
	return connect.NewResponse(&stubsv1.RemoveStubsResponse{}), nil
}

func stubsToProto(stubs []StubEntry) ([]*stubsv1.Stub, error) {
	pbStubs := []*stubsv1.Stub{}
	for _, stub := range stubs {
		pbStub := &stubsv1.Stub{}
		pbStub.SetRef(stubsv1.StubRef_builder{
			Id:     proto.String(stub.Key.ID),
			Target: proto.String(string(stub.Key.Name)),
		}.Build())
		if stub.ActiveIf != nil {
			pbStub.SetActiveIf(stub.ActiveIf.GetString())
		}
		if stub.Error != nil {
			pbStub.SetError(stub.Error.StubsError)
		}
		if stub.Message != nil {
			content, err := protojson.Marshal(stub.Message)
			if err != nil {
				return nil, err
			}
			pbStub.SetJson(string(content))
		}
		pbStubs = append(pbStubs, pbStub)
	}
	return pbStubs, nil
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

// Alias stubs error to not conflict with error interface
type StubsError = *stubsv1.Error

type StatusError struct {
	StubsError
}

func (s *StatusError) Error() string {
	return s.GetMessage()
}
