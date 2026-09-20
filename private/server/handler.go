package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectproto"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/sudorandom/fauxrpc"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"
	"golang.org/x/sync/errgroup"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// restProtocolName is the CallInfo.Protocol value vanguard sets for REST
// calls. Vanguard's REST codec decodes into plain proto messages, so the
// requestMessage fast path only applies to the other protocols.
const restProtocolName = "rest"

type methodHandler struct {
	method   protoreflect.MethodDescriptor
	faker    fauxrpc.ProtoFaker
	validate protovalidate.Validator
	server   Server
	logger   *fauxlog.Logger
	maxDepth int
}

// buildMethods returns one connect.Method per RPC of the service, each backed
// by a handler that serves stub or generated responses without generated
// code.
func buildMethods(
	sd protoreflect.ServiceDescriptor,
	faker fauxrpc.ProtoFaker,
	validate protovalidate.Validator,
	s Server,
	logger *fauxlog.Logger,
	maxDepth int,
) []connect.Method {
	methods := sd.Methods()
	out := make([]connect.Method, 0, methods.Len())
	for i := 0; i < methods.Len(); i++ {
		md := methods.Get(i)
		handler := &methodHandler{
			method:   md,
			faker:    faker,
			validate: validate,
			server:   s,
			logger:   logger,
			maxDepth: maxDepth,
		}
		out = append(out, connect.Method{
			Spec:    specForMethod(md),
			Handler: handler.handle,
		})
	}
	return out
}

func specForMethod(md protoreflect.MethodDescriptor) connect.Spec {
	streamType := connect.StreamTypeUnary
	if md.IsStreamingClient() {
		streamType |= connect.StreamTypeClient
	}
	if md.IsStreamingServer() {
		streamType |= connect.StreamTypeServer
	}
	var idempotency connect.IdempotencyLevel
	if opts, ok := md.Options().(*descriptorpb.MethodOptions); ok && opts != nil {
		idempotency = connect.IdempotencyLevel(opts.GetIdempotencyLevel())
	}
	return connect.Spec{
		Procedure:        fmt.Sprintf("/%s/%s", md.Parent().FullName(), md.Name()),
		StreamType:       streamType,
		IdempotencyLevel: idempotency,
		Schema:           md,
	}
}

func (h *methodHandler) handle(ctx context.Context, _ connect.Spec, stream connect.ServerStream) (retErr error) {
	startTime := time.Now()
	h.server.IncrementTotalRequests()
	info, _ := connect.CallInfoForServerContext(ctx)

	var requestBody releasableMessage
	var responseBody proto.Message
	var stubsUsed []fauxrpc.StubEntry
	reqFrameTracker := NewFrameTracker(10)
	resFrameTracker := NewFrameTracker(10)

	defer func() {
		h.logCall(startTime, info, requestBody, responseBody, stubsUsed, reqFrameTracker, resFrameTracker, retErr)
		if requestBody != nil {
			requestBody.Release()
		}
	}()

	var isFallback bool
	if h.server.GetProxyTo() != "" {
		err := h.handleProxy(ctx, info, stream, reqFrameTracker, resFrameTracker, &requestBody, &responseBody)
		if err == nil {
			return nil
		}
		if !isUnimplementedError(err) {
			h.server.IncrementErrors()
			return err
		}
		isFallback = true
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// Handle reading requests
	var input proto.Message
	switch {
	case isFallback:
		if requestBody != nil {
			input = requestBody
		}
	case h.method.IsStreamingClient():
		// Drain the request stream concurrently with the response; the
		// frames only feed the request log.
		eg.Go(func() error {
			for {
				msg, err := h.receive(info, stream)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				if err := h.validateMessage(msg); err != nil {
					msg.Release()
					return err
				}
				reqFrameTracker.Add(msg)
				msg.Release()
			}
		})
	default:
		msg, err := h.receive(info, stream)
		if err != nil && !errors.Is(err, io.EOF) {
			h.server.IncrementErrors()
			return err
		}
		if msg != nil {
			requestBody = msg
			input = msg
			if err := h.validateMessage(msg); err != nil {
				h.server.IncrementErrors()
				return err
			}
		}
	}

	// Handle writing the response
	eg.Go(func() error {
		stubFaker := stubs.NewStubFaker(h.server)
		celCtx := &protocel.CELContext{
			MethodDescriptor: h.method,
			Req:              input,
		}
		stubEntry, err := stubFaker.FindStub(egCtx, celCtx, h.method.Output())
		if err != nil {
			return connect.NewError(connect.CodeInternal, err.Error())
		}

		if stubEntry != nil && stubEntry.Stream != nil {
			stubsUsed = append(stubsUsed, stubEntry.Key)
			setFauxRPCHeaders(info, stubsUsed)
			return stubs.ExecuteStream(egCtx, stubEntry.Stream, h.method.Output(), celCtx, func(msg proto.Message) error {
				if err := stream.Send(msg); err != nil {
					return err
				}
				resFrameTracker.Add(msg)
				return nil
			}, nil)
		}

		out := registry.NewMessage(h.method.Output()).Interface()
		genOpts := fauxrpc.GenOptions{
			MaxDepth:     h.maxDepth,
			ViolateRules: h.server.GetViolateRules(),
			Context: protocel.WithCELContext(egCtx, &protocel.CELContext{
				MethodDescriptor: h.method,
				Req:              input,
			}),
			StubRecorder: func(stub fauxrpc.StubEntry) {
				stubsUsed = append(stubsUsed, stub)
			},
			// Where the extensions of each generated message are found.
			// Without it they stay unset, because a message descriptor
			// cannot name the extensions declared against it.
			Extensions: h.server.Types(),
		}
		if h.server.GetStaticSeed() {
			genOpts.Faker = gofakeit.New(staticSeedForMethod(h.method.FullName()))
		}
		if err := h.faker.SetDataOnMessage(out, genOpts); err != nil {
			var stubErr *stubs.StatusError
			switch {
			case errors.Is(err, fauxrpc.ErrNotFaked):
				// If we can't fake it, we should return the empty message instead of an error
				// This ensures the client gets a valid response structure
				slog.Warn("Failed to fake response data, returning empty message", "method", h.method.FullName(), "error", err)
			case errors.As(err, &stubErr):
				return connectErrorFromStub(stubErr.StubsError)
			default:
				return connect.NewError(connect.CodeInternal, err.Error())
			}
		}
		responseBody = out
		setFauxRPCHeaders(info, stubsUsed)
		if err := stream.Send(out); err != nil {
			return err
		}
		resFrameTracker.Add(out)
		return nil
	})

	if err := eg.Wait(); err != nil {
		h.server.IncrementErrors()
		var stubErr *stubs.StatusError
		if errors.As(err, &stubErr) {
			return connectErrorFromStub(stubErr.StubsError)
		}
		return err
	}
	return nil
}

// receive reads the next request message from the stream. Everything served
// by connecthttp goes through this package's codecs, which understand the
// requestMessage fast path; vanguard's REST codec needs a plain message.
func (h *methodHandler) receive(info *connect.CallInfo, stream connect.ServerStream) (releasableMessage, error) {
	if info != nil && info.Protocol != restProtocolName {
		req := &requestMessage{desc: h.method.Input()}
		if err := stream.Receive(req); err != nil {
			return nil, err
		}
		return req.msg, nil
	}
	msg := registry.NewMessage(h.method.Input()).Interface()
	if err := stream.Receive(msg); err != nil {
		return nil, err
	}
	return plainMessage{Message: msg}, nil
}

func (h *methodHandler) validateMessage(msg proto.Message) error {
	if h.validate == nil {
		return nil
	}
	err := h.validate.Validate(msg)
	if err == nil {
		return nil
	}
	connectErr := connect.NewError(connect.CodeInvalidArgument, err.Error())
	if validationErr := new(protovalidate.ValidationError); errors.As(err, &validationErr) {
		if detail, detailErr := connectproto.NewErrorDetail(validationErr.ToProto()); detailErr == nil {
			connectErr = connectErr.WithDetail(detail)
		} else {
			slog.Error("error serializing validation details", "error", detailErr)
		}
	}
	return connectErr
}

func (h *methodHandler) logCall(
	startTime time.Time,
	info *connect.CallInfo,
	requestBody releasableMessage,
	responseBody proto.Message,
	stubsUsed []fauxrpc.StubEntry,
	reqFrameTracker, resFrameTracker *FrameTracker,
	retErr error,
) {
	duration := time.Since(startTime)

	clientProtocol := "unknown"
	var reqHeaders, resHeaders json.RawMessage
	if info != nil {
		clientProtocol = displayProtocol(info.Protocol)
		reqHeaders, _ = json.Marshal(maskHeaders(httpHeaderFromConnect(info.RequestHeader())))
		merged := httpHeaderFromConnect(info.ResponseHeader())
		for key, values := range info.ResponseTrailer().All() {
			for _, value := range values {
				merged.Add(key, value)
			}
		}
		resHeaders, _ = json.Marshal(merged)
	}

	var reqBodyBytes []byte
	if requestBody != nil {
		reqBodyBytes, _ = protojson.Marshal(requestBody)
	}
	var resBodyBytes []byte
	if responseBody != nil {
		resBodyBytes, _ = protojson.Marshal(responseBody)
	}

	code := 0 // OK
	if retErr != nil {
		connectErr := asConnectError(retErr)
		code = int(connectErr.Code())
		st := &statuspb.Status{
			Code:    int32(code), //nolint:gosec // connect codes are small
			Message: connectErr.Message(),
		}
		for _, detail := range connectErr.Details() {
			st.Details = append(st.Details, connectproto.ErrorDetailToAny(detail))
		}
		if jsonBytes, err := protojson.Marshal(st); err == nil {
			resBodyBytes = jsonBytes
		}
	}

	h.logger.Log(&fauxlog.LogEntry{
		ID:              uuid.New().String(),
		Timestamp:       startTime,
		Service:         string(h.method.Parent().FullName()),
		Method:          string(h.method.Name()),
		ClientProtocol:  clientProtocol,
		Status:          code,
		Duration:        duration,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: resHeaders,
		RequestBody:     reqBodyBytes,
		ResponseBody:    resBodyBytes,
		RequestFrames:   reqFrameTracker.Frames(),
		ResponseFrames:  resFrameTracker.Frames(),
		StubsUsed:       stubsUsed,
	})
}

func displayProtocol(protocol string) string {
	switch protocol {
	case connect.ProtocolNameGRPC:
		return "gRPC"
	case connect.ProtocolNameGRPCWeb:
		return "gRPC-Web"
	case connect.ProtocolNameConnect:
		return "ConnectRPC"
	case restProtocolName:
		return "HTTP"
	default:
		return "unknown"
	}
}

func httpHeaderFromConnect(header *connect.Header) http.Header {
	out := make(http.Header)
	if header == nil {
		return out
	}
	for key, values := range header.All() {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}

func asConnectError(err error) *connect.Error {
	connectErr := new(connect.Error)
	if errors.As(err, &connectErr) {
		return connectErr
	}
	return connect.NewError(connect.CodeUnknown, err.Error())
}

func connectErrorFromStub(e *stubsv1.Error) *connect.Error {
	connectErr := connect.NewError(connect.Code(e.GetCode()), e.GetMessage())
	for _, detail := range e.GetDetails() {
		connectErr = connectErr.WithDetail(&connect.ErrorDetail{
			Type:  strings.TrimPrefix(detail.GetTypeUrl(), "type.googleapis.com/"),
			Value: detail.GetValue(),
		})
	}
	return connectErr
}

func staticSeedForMethod(method protoreflect.FullName) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(method))
	return hasher.Sum64()
}

func setFauxRPCHeaders(info *connect.CallInfo, stubsUsed []fauxrpc.StubEntry) {
	if info == nil {
		return
	}
	header := info.ResponseHeader()
	if len(stubsUsed) > 0 {
		header.Set("x-fauxrpc-source", "stub")
		var ids []string
		for _, stub := range stubsUsed {
			if id := stub.GetID(); id != "" {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			header.Set("x-fauxrpc-mock-ids", strings.Join(ids, ", "))
		}
	} else {
		header.Set("x-fauxrpc-source", "fake")
	}
}
