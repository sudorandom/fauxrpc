package server

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"strings"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/sudorandom/fauxrpc/private/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

type proxyTransport struct {
	httpTransport  http.RoundTripper
	httpsTransport http.RoundTripper
}

func (t *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "https" {
		return t.httpsTransport.RoundTrip(req)
	}
	return t.httpTransport.RoundTrip(req)
}

func newProxyClient() *http.Client {
	httpTrans := &http.Transport{}
	httpTrans.Protocols = new(http.Protocols)
	httpTrans.Protocols.SetUnencryptedHTTP2(true)

	httpsTrans := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpsTrans.Protocols = new(http.Protocols)
	httpsTrans.Protocols.SetHTTP1(true)
	httpsTrans.Protocols.SetHTTP2(true)
	return &http.Client{
		Transport: &proxyTransport{
			httpTransport:  httpTrans,
			httpsTransport: httpsTrans,
		},
	}
}

// handleProxy relays the call to the upstream configured with ProxyTo over
// gRPC, mirroring headers and frames in both directions. A CodeUnimplemented
// error from upstream is returned so the caller can fall back to a generated
// response.
func (h *methodHandler) handleProxy(
	ctx context.Context,
	info *connect.CallInfo,
	stream connect.ServerStream,
	reqFrameTracker, resFrameTracker *FrameTracker,
	requestBody *releasableMessage,
	responseBody *proto.Message,
) error {
	upstream := h.server.GetProxyTo()
	if !strings.HasPrefix(upstream, "http://") && !strings.HasPrefix(upstream, "https://") {
		upstream = "http://" + upstream
	}
	upstream = strings.TrimSuffix(upstream, "/")

	transport := connecthttp.NewTransport(h.server.GetProxyClient(), upstream, connecthttp.WithGRPC())
	client := connect.NewClient(transport)
	spec := specForMethod(h.method)

	callCtx, callInfo := connect.NewClientContext(ctx)
	if info != nil {
		copyFilteredHeaders(info.RequestHeader(), callInfo.RequestHeader())
	}

	relayResponseHeaders := func() {
		if info == nil {
			return
		}
		copyFilteredHeaders(callInfo.ResponseHeader(), info.ResponseHeader())
		info.ResponseHeader().Set("x-fauxrpc-source", "proxy")
	}

	isClientStream := h.method.IsStreamingClient()
	isServerStream := h.method.IsStreamingServer()

	switch {
	case !isClientStream && !isServerStream:
		reqMsg, err := h.receive(info, stream)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		var req proto.Message
		if reqMsg != nil {
			*requestBody = reqMsg
			req = reqMsg
		} else {
			req = registry.NewMessage(h.method.Input()).Interface()
		}

		res := dynamicpb.NewMessage(h.method.Output())
		if err := client.CallUnary(callCtx, spec, req, res); err != nil {
			return err
		}
		relayResponseHeaders()
		*responseBody = res
		return stream.Send(res)

	case !isClientStream && isServerStream:
		reqMsg, err := h.receive(info, stream)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		var req proto.Message
		if reqMsg != nil {
			*requestBody = reqMsg
			req = reqMsg
		} else {
			req = registry.NewMessage(h.method.Input()).Interface()
		}

		upstreamStream, err := client.CallServerStream(callCtx, spec, req)
		if err != nil {
			return err
		}
		defer func() { _ = upstreamStream.Close() }()

		headersRelayed := false
		for {
			res := dynamicpb.NewMessage(h.method.Output())
			if err := upstreamStream.Receive(res); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			if !headersRelayed {
				relayResponseHeaders()
				headersRelayed = true
			}
			resFrameTracker.Add(res)
			if err := stream.Send(res); err != nil {
				return err
			}
		}
		if !headersRelayed {
			relayResponseHeaders()
		}
		return nil

	case isClientStream && !isServerStream:
		upstreamStream, err := client.CallClientStream(callCtx, spec)
		if err != nil {
			return err
		}
		defer func() { _ = upstreamStream.Close() }()

		for {
			msg, err := h.receive(info, stream)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			reqFrameTracker.Add(msg)
			if err := upstreamStream.Send(msg); err != nil {
				msg.Release()
				return err
			}
			msg.Release()
		}
		if err := upstreamStream.CloseSend(); err != nil {
			return err
		}

		res := dynamicpb.NewMessage(h.method.Output())
		if err := upstreamStream.Receive(res); err != nil {
			return err
		}
		relayResponseHeaders()
		*responseBody = res
		return stream.Send(res)

	default: // bidi
		upstreamStream, err := client.CallClientStream(callCtx, spec)
		if err != nil {
			return err
		}
		defer func() { _ = upstreamStream.Close() }()

		var eg errgroup.Group
		eg.Go(func() error {
			for {
				msg, err := h.receive(info, stream)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return upstreamStream.CloseSend()
					}
					return err
				}
				reqFrameTracker.Add(msg)
				if err := upstreamStream.Send(msg); err != nil {
					msg.Release()
					return err
				}
				msg.Release()
			}
		})
		eg.Go(func() error {
			headersRelayed := false
			for {
				res := dynamicpb.NewMessage(h.method.Output())
				if err := upstreamStream.Receive(res); err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				if !headersRelayed {
					relayResponseHeaders()
					headersRelayed = true
				}
				resFrameTracker.Add(res)
				if err := stream.Send(res); err != nil {
					return err
				}
			}
		})
		return eg.Wait()
	}
}

// copyFilteredHeaders copies user headers between connect header maps,
// skipping protocol-managed keys.
func copyFilteredHeaders(src, dst *connect.Header) {
	if src == nil || dst == nil {
		return
	}
	for key, values := range src.All() {
		keyLower := strings.ToLower(key)
		if strings.HasPrefix(keyLower, "content-") ||
			strings.HasPrefix(keyLower, "grpc-") ||
			strings.HasPrefix(keyLower, "connect-") ||
			keyLower == "connection" ||
			keyLower == "te" ||
			keyLower == "trailer" ||
			keyLower == "host" ||
			keyLower == "accept-encoding" {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func isUnimplementedError(err error) bool {
	if err == nil {
		return false
	}
	var connectErr *connect.Error
	if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeUnimplemented {
		return true
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
		return true
	}
	return false
}
