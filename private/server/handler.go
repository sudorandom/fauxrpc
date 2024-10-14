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
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/grpc"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	stubsv1 "github.com/sudorandom/fauxrpc/proto/gen/stubs/v1"
)

func NewHandler(service protoreflect.ServiceDescriptor, db stubs.StubDatabase) http.Handler {
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

		// completely ignore the body. Maybe later we'll need it as input to the response message
		go func() {
			_, _ = io.Copy(io.Discard, r.Body)
		}()

		slog.Info("MethodCalled", slog.String("service", serviceName), slog.String("method", methodName))

		out, err := fauxrpc.NewMessage(method.Output(), fauxrpc.GenOptions{MaxDepth: 20, StubDB: db})
		if err != nil {
			var statusErr *stubs.StatusError
			if errors.As(err, &statusErr) {
				status := grpcStatusFromError(statusErr.StubsError)
				w.Header().Set("Grpc-Status", strconv.FormatUint(uint64(status.Code), 10))
				w.Header().Set("Grpc-Message", status.Message)
				var bin []byte
				if len(status.Details) > 0 {
					bin, err = proto.Marshal(status)
					slog.Warn("failed to marshal grpc-status-details-bin", "error", err)
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
	return &status.Status{
		Code:    int32(e.Code),
		Message: e.GetMessage(),
		Details: e.Details,
	}
}
