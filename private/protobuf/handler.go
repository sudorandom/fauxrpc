package protobuf

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewHandler(service protoreflect.ServiceDescriptor) http.Handler {
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

		// completely ignore the body. Maybe later we'll need it as input to the response message
		go func() {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}()

		slog.Info("MethodCalled", slog.String("service", serviceName), slog.String("method", methodName))

		out := fauxrpc.NewMessage(method.Output())

		b, err := proto.Marshal(out)
		if err != nil {
			slog.Error(fmt.Sprintf("error marshalling msg: %s", err))
			w.Header().Set("Grpc-Status", connect.CodeInternal.String())
			w.Header().Set("Grpc-Message", err.Error())
			return
		}
		writeGRPCMessage(w, b)
		w.Header().Set("Grpc-Status", "0")
		w.Header().Set("Grpc-Message", "")
	})
}
