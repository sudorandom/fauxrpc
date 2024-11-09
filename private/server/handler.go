package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/grpc"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/dynamicpb"

	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
)

const maxMessageSize = 4 * 1024 * 1024

func NewHandler(service protoreflect.ServiceDescriptor, db stubs.StubDatabase, validate *protovalidate.Validator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trailer", "Grpc-Status,Grpc-Message,Grpc-Status-Details-Bin")
		w.Header().Add("Content-Type", "application/grpc")

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			grpcWriteStatus(w, status.New(codes.NotFound, ""))
			return
		}

		serviceName := parts[1]
		if serviceName != string(service.FullName()) {
			grpcWriteStatus(w, status.New(codes.NotFound, "service not found"))
			return
		}
		methodName := parts[2]
		method := service.Methods().ByName(protoreflect.Name(methodName))
		if method == nil {
			grpcWriteStatus(w, status.New(codes.NotFound, "method not found"))
			return
		}
		defer r.Body.Close()

		readMessage := func() *status.Status {
			body := make([]byte, maxMessageSize)
			size, err := grpc.ReadGRPCMessage(r.Body, body)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return status.New(codes.NotFound, err.Error())
			}
			msg := newMessage(method.Input()).Interface()
			if err := proto.Unmarshal(body[:size], msg); err != nil {
				return status.New(codes.NotFound, err.Error())
			}
			if err := validate.Validate(msg); err != nil {
				grpcErr := status.New(codes.InvalidArgument, err.Error())
				if validationErr := new(protovalidate.ValidationError); errors.As(err, &validationErr) {
					grpcErr, err = grpcErr.WithDetails(validationErr.ToProto())
					if err != nil {
						slog.Error("error serializing validation details", "error", err)
					}
				}
				return grpcErr
			}
			return nil
		}

		slog.Info("MethodCalled", slog.String("service", serviceName), slog.String("method", methodName))

		eg, _ := errgroup.WithContext(r.Context())

		// Handle reading requests
		if validate == nil {
			eg.Go(func() error {
				_, _ = io.Copy(io.Discard, r.Body)
				return nil
			})
		} else {
			if method.IsStreamingClient() {
				// completely ignore the body. Maybe later we'll need it as input to the response message
				eg.Go(func() error {
					for {
						if st := readMessage(); st != nil {
							return st.Err()
						}
					}
				})
			} else {
				if st := readMessage(); st != nil {
					grpcWriteStatus(w, st)
					return
				}
			}
		}

		// Handle writing response
		var msg []byte
		eg.Go(func() error {
			out, err := fauxrpc.NewMessage(method.Output(), fauxrpc.GenOptions{MaxDepth: 20, StubDB: db})
			if err != nil {
				var statusErr *stubs.StatusError
				if errors.As(err, &statusErr) {
					return status.New(codes.Code(statusErr.Code), statusErr.Error()).Err()
				}
				return status.New(codes.Internal, err.Error()).Err()
			}

			b, err := proto.Marshal(out)
			if err != nil {
				slog.Error(fmt.Sprintf("error marshalling msg: %s", err))
				return status.New(codes.Internal, err.Error()).Err()
			}
			msg = b
			return nil
		})

		// Write response
		if err := eg.Wait(); err != nil {
			if st, ok := status.FromError(err); ok {
				grpcWriteStatus(w, st)
				return
			} else {
				grpcWriteStatus(w, status.New(codes.Internal, err.Error()))
				return
			}
		}
		grpc.WriteGRPCMessage(w, msg)
		grpcWriteStatus(w, status.New(codes.OK, ""))
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
	status := status.New(codes.Code(e.Code), e.GetMessage())
	if len(e.Details) > 0 {
		details := make([]protoiface.MessageV1, len(e.Details))
		for i, detail := range e.Details {
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

func newMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return dynamicpb.NewMessageType(md).New()
	}
	return mt.New()
}
