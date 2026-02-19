package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"buf.build/go/protovalidate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/sudorandom/fauxrpc"
	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
	"github.com/sudorandom/fauxrpc/private/grpc"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"
	"golang.org/x/sync/errgroup"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

const maxMessageSize = 4 * 1024 * 1024

var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, maxMessageSize)
		return &b
	},
}

func NewHandler(service protoreflect.ServiceDescriptor, faker fauxrpc.ProtoFaker, validate protovalidate.Validator, s Server, logger *fauxlog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		s.IncrementTotalRequests()

		var finalStatus *status.Status
		var requestBody proto.Message
		var responseBody proto.Message
		var stubsUsed []fauxrpc.StubEntry
		reqFrameTracker := NewFrameTracker(10)
		resFrameTracker := NewFrameTracker(10)

		parts := strings.Split(r.URL.Path, "/")
		var serviceName, methodName string
		if len(parts) == 3 {
			serviceName = parts[1]
			methodName = parts[2]
		}

		defer func() {
			duration := time.Since(startTime)

			clientProtocol := "unknown"
			if protocol, ok := r.Context().Value(clientProtocolKey).(string); ok {
				clientProtocol = protocol
			}

			var reqHeaders json.RawMessage
			if headers, ok := r.Context().Value(requestHeadersKey).([]byte); ok {
				reqHeaders = headers
			}

			resHeaders, _ := json.Marshal(w.Header())

			var reqBodyBytes []byte
			if requestBody != nil {
				reqBodyBytes, _ = protojson.Marshal(requestBody)
			}

			var resBodyBytes []byte
			if responseBody != nil {
				resBodyBytes, _ = protojson.Marshal(responseBody)
			}

			code := codes.Unknown
			if finalStatus != nil {
				code = finalStatus.Code()
			}

			if code != codes.OK {
				if statusDetailsBin := w.Header().Get("Grpc-Status-Details-Bin"); statusDetailsBin != "" {
					decoded, err := base64.StdEncoding.DecodeString(statusDetailsBin)
					if err == nil {
						st := &statuspb.Status{}
						if err := proto.Unmarshal(decoded, st); err == nil {
							jsonBytes, err := protojson.Marshal(st)
							if err == nil {
								resBodyBytes = jsonBytes
							}
						}
					}
				}
			}
			logger.Log(&fauxlog.LogEntry{
				ID:              uuid.New().String(),
				Timestamp:       startTime,
				Service:         serviceName,
				Method:          methodName,
				ClientProtocol:  clientProtocol,
				Status:          int(code),
				Duration:        duration,
				RequestHeaders:  reqHeaders,
				ResponseHeaders: resHeaders,
				RequestBody:     reqBodyBytes,
				ResponseBody:    resBodyBytes,
				RequestFrames:   reqFrameTracker.Frames(),
				ResponseFrames:  resFrameTracker.Frames(),
				StubsUsed:       stubsUsed,
			})
		}()

		w.Header().Set("Trailer", "Grpc-Status,Grpc-Message,Grpc-Status-Details-Bin")
		w.Header().Add("Content-Type", "application/grpc")

		if len(parts) != 3 {
			s.IncrementErrors()
			finalStatus = status.New(codes.NotFound, "")
			grpcWriteStatus(w, finalStatus)
			return
		}

		if serviceName != string(service.FullName()) {
			s.IncrementErrors()
			finalStatus = status.New(codes.NotFound, "service not found")
			grpcWriteStatus(w, finalStatus)
			return
		}
		method := service.Methods().ByName(protoreflect.Name(methodName))
		if method == nil {
			s.IncrementErrors()
			finalStatus = status.New(codes.NotFound, "method not found")
			grpcWriteStatus(w, finalStatus)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		readMessageBuf := bufferPool.Get().(*[]byte)
		defer bufferPool.Put(readMessageBuf)

		readMessage := func() (proto.Message, *status.Status) {
			size, err := grpc.ReadGRPCMessage(r.Body, *readMessageBuf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil, nil
				}
				s.IncrementErrors()
				return nil, status.New(codes.NotFound, err.Error())
			}
			msg := registry.NewMessage(method.Input()).Interface()
			if err := proto.Unmarshal((*readMessageBuf)[:size], msg); err != nil {
				s.IncrementErrors()
				return nil, status.New(codes.NotFound, err.Error())
			}
			if err := validate.Validate(msg); err != nil {
				s.IncrementErrors()
				grpcErr := status.New(codes.InvalidArgument, err.Error())
				if validationErr := new(protovalidate.ValidationError); errors.As(err, &validationErr) {
					grpcErr, err = grpcErr.WithDetails(validationErr.ToProto())
					if err != nil {
						slog.Error("error serializing validation details", "error", err)
					}
				}
				return msg, grpcErr
			}
			return msg, nil
		}

		eg, ctx := errgroup.WithContext(r.Context())

		// Handle reading requests
		var input proto.Message
		if method.IsStreamingClient() {
			// completely ignore the body. Maybe later we'll need it as input to the response message
			eg.Go(func() error {
				for {
					msg, st := readMessage()
					if st != nil {
						return st.Err()
					}
					if msg == nil {
						return nil
					}
					reqFrameTracker.Add(msg)
				}
			})
		} else {
			var st *status.Status
			input, st = readMessage()
			requestBody = input
			if st != nil {
				s.IncrementErrors()
				finalStatus = st
				grpcWriteStatus(w, st)
				return
			}
		}

		// Handle writing response
		eg.Go(func() error {
			stubFaker := stubs.NewStubFaker(s)
			celCtx := &protocel.CELContext{
				MethodDescriptor: method,
				Req:              input,
			}
			stubEntry, err := stubFaker.FindStub(ctx, celCtx, method.Output())
			if err != nil {
				return status.New(codes.Internal, err.Error()).Err()
			}

			if stubEntry != nil && stubEntry.Stream != nil {
				stubsUsed = append(stubsUsed, stubEntry.Key)
				return stubs.ExecuteStream(ctx, stubEntry.Stream, method.Output(), celCtx, func(msg proto.Message) error {
					b, err := proto.Marshal(msg)
					if err != nil {
						return status.New(codes.Internal, err.Error()).Err()
					}
					if err := grpc.WriteGRPCMessage(w, b); err != nil {
						return err
					}
					resFrameTracker.Add(msg)
					return nil
				}, nil)
			}

			out := registry.NewMessage(method.Output()).Interface()
			genOpts := fauxrpc.GenOptions{
				MaxDepth: 20,
				Faker:    gofakeit.New(0),
				Context: protocel.WithCELContext(ctx, &protocel.CELContext{
					MethodDescriptor: method,
					Req:              input,
				}),
				StubRecorder: func(stub fauxrpc.StubEntry) {
					stubsUsed = append(stubsUsed, stub)
				},
			}
			if err := faker.SetDataOnMessage(out, genOpts); err != nil {
				var stubErr *stubs.StatusError
				s.IncrementErrors()
				switch {
				case errors.Is(err, fauxrpc.ErrNotFaked):
					// If we can't fake it, we should return the empty message instead of an error
					// This ensures the client gets a valid response structure
					slog.Warn("Failed to fake response data, returning empty message", "method", method.FullName(), "error", err)
				case errors.As(err, &stubErr):
					return grpcStatusFromError(stubErr.StubsError).Err()
				default:
					return status.New(codes.Internal, err.Error()).Err()
				}
			}
			responseBody = out

			b, err := proto.Marshal(out)
			if err != nil {
				s.IncrementErrors()
				slog.Error(fmt.Sprintf("error marshalling msg: %s", err))
				return status.New(codes.Internal, err.Error()).Err()
			}
			return grpc.WriteGRPCMessage(w, b)
		})

		// Write response
		if err := eg.Wait(); err != nil {
			s.IncrementErrors()
			var stubErr *stubs.StatusError
			if errors.As(err, &stubErr) {
				finalStatus = grpcStatusFromError(stubErr.StubsError)
				grpcWriteStatus(w, finalStatus)
				return
			} else if st, ok := status.FromError(err); ok {
				finalStatus = st
				grpcWriteStatus(w, st)
				return
			} else {
				finalStatus = status.New(codes.Internal, err.Error())
				grpcWriteStatus(w, finalStatus)
				return
			}
		}
		finalStatus = status.New(codes.OK, "")
		grpcWriteStatus(w, finalStatus)
	})
}

func grpcWriteStatus(w http.ResponseWriter, st *status.Status) {
	w.Header().Set("Grpc-Status", strconv.FormatInt(int64(st.Code()), 10))
	w.Header().Set("Grpc-Message", st.Message())
	if details, err := proto.Marshal(st.Proto()); err != nil {
		slog.Error("error serializing validation details", "error", err)
	} else {
		w.Header().Set("Grpc-Status-Details-Bin", base64.StdEncoding.EncodeToString(details))
	}
}

func grpcStatusFromError(e *stubsv1.Error) *status.Status {
	status := status.New(codes.Code(e.GetCode()), e.GetMessage())
	if len(e.GetDetails()) > 0 {
		details := make([]protoiface.MessageV1, len(e.GetDetails()))
		for i, detail := range e.GetDetails() {
			details[i] = detail
		}
		s, err := status.WithDetails(details...)
		if err != nil {
			slog.Warn("unable to add details to status", "error", err)
		} else {
			status = s
		}
	}
	return status
}
