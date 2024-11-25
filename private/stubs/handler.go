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
	entries := make([]StubEntry, len(req.Msg.Stubs))
	stubs := make([]*stubsv1.Stub, len(req.Msg.Stubs))
	for i, stub := range req.Msg.Stubs {
		if stub.Ref == nil {
			stub.Ref = &stubsv1.StubRef{}
		}
		if stub.GetRef().Id == "" {
			stub.Ref.Id = gofakeit.AdjectiveDescriptive() + "-" + strings.ReplaceAll(gofakeit.Animal(), " ", "-") + gofakeit.DigitN(3)
		}

		ref := stub.GetRef()
		name, err := normalizeTargetName(ref.GetTarget())
		if err != nil {
			return nil, err
		}

		entry := StubEntry{
			Priority: int(stub.GetPriority()),
		}

		desc, err := h.registry.Files().FindDescriptorByName(name)
		if err != nil {
			return nil, fmt.Errorf("unable to find object named %s: %w", name, err)
		}
		var md protoreflect.MessageDescriptor
		switch t := desc.(type) {
		case protoreflect.MethodDescriptor:
			if len(stub.ActiveIf) > 0 {
				r, err := NewActiveIf(t, stub.ActiveIf)
				if err != nil {
					return nil, err
				}
				entry.ActiveIf = r
			}

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
		entry.Key = StubKey{ID: stub.GetRef().GetId(), Name: name}

		switch t := stub.GetContent().(type) {
		case *stubsv1.Stub_Json:
			msg := registry.NewMessage(md).Interface()
			if err := protojson.Unmarshal([]byte(t.Json), msg); err != nil {
				return nil, err
			}
			entry.Message = msg
		case *stubsv1.Stub_Proto:
			msg := registry.NewMessage(md).Interface()
			if err := proto.Unmarshal(t.Proto, msg); err != nil {
				return nil, err
			}
			entry.Message = msg
		case *stubsv1.Stub_Error:
			entry.Error = &StatusError{StubsError: t.Error}
		}

		entries[i] = entry
		stubs[i] = stub
	}

	for _, entry := range entries {
		h.stubdb.AddStub(entry)
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
	return connect.NewResponse(&stubsv1.ListStubsResponse{Stubs: pbstubs}), nil

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
		pbStub := &stubsv1.Stub{
			Ref: &stubsv1.StubRef{
				Id:     stub.Key.ID,
				Target: string(stub.Key.Name),
			},
		}
		if stub.ActiveIf != nil {
			pbStub.ActiveIf = stub.ActiveIf.GetString()
		}
		if stub.Error != nil {
			pbStub.Content = &stubsv1.Stub_Error{Error: stub.Error.StubsError}
		}
		if stub.Message != nil {
			content, err := protojson.Marshal(stub.Message)
			if err != nil {
				return nil, err
			}
			pbStub.Content = &stubsv1.Stub_Json{Json: string(content)}
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
	return s.Message
}
