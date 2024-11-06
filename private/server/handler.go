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

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/grpc"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/dynamicpb"

	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
)

const maxMessageSize = 2 * 1024 * 1024 * 1024

func NewHandler(service protoreflect.ServiceDescriptor, db stubs.StubDatabase, validate *protovalidate.Validator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trailer", "Grpc-Status,Grpc-Message")
		w.Header().Add("Content-Type", "application/grpc")

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			w.Header().Set("Grpc-Status", "5")
			w.Header().Set("Grpc-Message", "")
			return
		}

		serviceName := parts[1]
		if serviceName != string(service.FullName()) {
			w.Header().Set("Grpc-Status", "5")
			w.Header().Set("Grpc-Message", "service not found")
			return
		}
		methodName := parts[2]
		method := service.Methods().ByName(protoreflect.Name(methodName))
		if method == nil {
			w.Header().Set("Grpc-Status", "5")
			w.Header().Set("Grpc-Message", "method not found")
			return
		}
		defer r.Body.Close()

		if method.IsStreamingClient() || validate == nil {
			// completely ignore the body. Maybe later we'll need it as input to the response message
			go func() {
				_, _ = io.Copy(io.Discard, r.Body)
			}()
		} else {
			body := make([]byte, maxMessageSize)
			size, err := grpc.ReadGRPCMessage(r.Body, body)
			if err != nil {
				w.Header().Set("Grpc-Status", "5")
				w.Header().Set("Grpc-Message", fmt.Sprintf("invalid protobuf message received: %s", err))
				return
			}
			msg := newMessage(method.Input()).Interface()
			if err := proto.Unmarshal(body[:size], msg); err != nil {
				w.Header().Set("Grpc-Status", "5")
				w.Header().Set("Grpc-Message", err.Error())
				return
			}
			if err := validate.Validate(msg); err != nil {
				w.Header().Set("Grpc-Status", "3")
				w.Header().Set("Grpc-Message", err.Error())
				grpcErr := status.New(codes.InvalidArgument, err.Error())
				if validationErr := new(protovalidate.ValidationError); errors.As(err, &validationErr) {
					grpcErr, err = grpcErr.WithDetails(validationErr.ToProto())
					if err != nil {
						slog.Error("error serializing validation details", "error", err)
					}
				}
				if details, err := proto.Marshal(grpcErr.Proto()); err != nil {
					slog.Error("error serializing validation details", "error", err)
				} else {
					w.Header().Set("Grpc-Status-Details-Bin", base64.StdEncoding.EncodeToString(details))
				}
				return
			}
		}

		slog.Info("MethodCalled", slog.String("service", serviceName), slog.String("method", methodName))

		out, err := fauxrpc.NewMessage(method.Output(), fauxrpc.GenOptions{MaxDepth: 20, StubDB: db})
		if err != nil {
			var statusErr *stubs.StatusError
			if errors.As(err, &statusErr) {
				statusErr := grpcStatusFromError(statusErr.StubsError)
				w.Header().Set("Grpc-Status", strconv.FormatUint(uint64(statusErr.Code()), 10))
				w.Header().Set("Grpc-Message", statusErr.Message())
				var bin []byte
				if len(statusErr.Details()) > 0 {
					bin, err = proto.Marshal(statusErr.Proto())
					if err != nil {
						slog.Warn("failed to marshal grpc-status-details-bin", "error", err)
					}
				}
				w.Header().Set("Grpc-Status-Details-Bin", base64.RawStdEncoding.EncodeToString(bin))
				return
			}

			w.Header().Set("Grpc-Status", "13")
			w.Header().Set("Grpc-Message", err.Error())
		}

		b, err := proto.Marshal(out)
		if err != nil {
			slog.Error(fmt.Sprintf("error marshalling msg: %s", err))
			w.Header().Set("Grpc-Status", connect.CodeInternal.String())
			w.Header().Set("Grpc-Message", err.Error())
			return
		}
		grpc.WriteGRPCMessage(w, b)
		w.Header().Set("Grpc-Status", "0")
		w.Header().Set("Grpc-Message", "")
	})
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
