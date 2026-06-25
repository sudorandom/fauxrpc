package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sudorandom/fauxrpc/private/grpc"
	"golang.org/x/net/http2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	_ = http2.ConfigureTransport(httpsTrans)
	return &http.Client{
		Transport: &proxyTransport{
			httpTransport:  httpTrans,
			httpsTransport: httpsTrans,
		},
	}
}

type dynamicProtoCodec struct {
	methodDesc protoreflect.MethodDescriptor
}

func (c *dynamicProtoCodec) Name() string {
	return "proto"
}

func (c *dynamicProtoCodec) Marshal(msg any) ([]byte, error) {
	switch m := msg.(type) {
	case proto.Message:
		return proto.Marshal(m)
	case **dynamicpb.Message:
		if *m == nil {
			return nil, fmt.Errorf("cannot marshal nil **dynamicpb.Message")
		}
		return proto.Marshal(*m)
	case dynamicpb.Message:
		return proto.Marshal(&m)
	default:
		return nil, fmt.Errorf("can't marshal %T", msg)
	}
}

func (c *dynamicProtoCodec) Unmarshal(binary []byte, msg any) error {
	if ptr, ok := msg.(**dynamicpb.Message); ok {
		newMsg := dynamicpb.NewMessage(c.methodDesc.Output())
		if err := proto.Unmarshal(binary, newMsg); err != nil {
			return err
		}
		*ptr = newMsg
		return nil
	}
	if m, ok := msg.(*dynamicpb.Message); ok {
		newMsg := dynamicpb.NewMessage(c.methodDesc.Output())
		if err := proto.Unmarshal(binary, newMsg); err != nil {
			return err
		}
		*m = *newMsg
		return nil
	}
	p, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("can't unmarshal into %T", msg)
	}
	return proto.Unmarshal(binary, p)
}

func handleProxy(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	s Server,
	method protoreflect.MethodDescriptor,
	serviceName, methodName string,
	reqFrameTracker, resFrameTracker *FrameTracker,
	requestBody *releasableMessage,
	responseBody *proto.Message,
) error {
	upstream := s.GetProxyTo()
	if !strings.HasPrefix(upstream, "http://") && !strings.HasPrefix(upstream, "https://") {
		upstream = "http://" + upstream
	}
	upstream = strings.TrimSuffix(upstream, "/")

	client := connect.NewClient[dynamicpb.Message, dynamicpb.Message](
		s.GetProxyClient(),
		upstream+"/"+serviceName+"/"+methodName,
		connect.WithGRPC(),
		connect.WithCodec(&dynamicProtoCodec{methodDesc: method}),
	)

	isClientStream := method.IsStreamingClient()
	isServerStream := method.IsStreamingServer()

	copyHeaders := func(src http.Header, dst http.Header) {
		for k, vv := range src {
			kl := strings.ToLower(k)
			if strings.HasPrefix(kl, "content-") ||
				strings.HasPrefix(kl, "grpc-") ||
				strings.HasPrefix(kl, "connect-") ||
				kl == "connection" ||
				kl == "te" ||
				kl == "trailer" ||
				kl == "host" ||
				kl == "accept-encoding" {
				continue
			}
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}

	if !isClientStream && !isServerStream {
		reqMsg, st := readUnaryRequest(r, method.Input())
		if st != nil {
			return st.Err()
		}
		*requestBody = reqMsg

		reqBytes, err := proto.Marshal(reqMsg)
		if err != nil {
			return err
		}
		dynamicReq := dynamicpb.NewMessage(method.Input())
		if err := proto.Unmarshal(reqBytes, dynamicReq); err != nil {
			return err
		}
		req := connect.NewRequest(dynamicReq)
		copyHeaders(r.Header, req.Header())

		resp, err := client.CallUnary(ctx, req)
		if err != nil {
			return err
		}

		copyHeaders(resp.Header(), w.Header())
		w.Header().Set("x-fauxrpc-source", "proxy")
		*responseBody = resp.Msg

		b, err := proto.Marshal(resp.Msg)
		if err != nil {
			return err
		}
		return grpc.WriteGRPCMessage(w, b)
	}

	if !isClientStream && isServerStream {
		reqMsg, st := readUnaryRequest(r, method.Input())
		if st != nil {
			return st.Err()
		}
		*requestBody = reqMsg

		reqBytes, err := proto.Marshal(reqMsg)
		if err != nil {
			return err
		}
		dynamicReq := dynamicpb.NewMessage(method.Input())
		if err := proto.Unmarshal(reqBytes, dynamicReq); err != nil {
			return err
		}
		req := connect.NewRequest(dynamicReq)
		copyHeaders(r.Header, req.Header())

		stream, err := client.CallServerStream(ctx, req)
		if err != nil {
			return err
		}

		copyHeaders(stream.ResponseHeader(), w.Header())
		w.Header().Set("x-fauxrpc-source", "proxy")

		for stream.Receive() {
			respMsg := stream.Msg()
			resFrameTracker.Add(respMsg)

			respBytes, err := proto.Marshal(respMsg)
			if err != nil {
				return err
			}
			if err := grpc.WriteGRPCMessage(w, respBytes); err != nil {
				return err
			}
		}

		if err := stream.Err(); err != nil {
			return err
		}

		return nil
	}

	if isClientStream && !isServerStream {
		stream := client.CallClientStream(ctx)
		copyHeaders(r.Header, stream.RequestHeader())

		var firstReq proto.Message

		readMessageBuf := bufferPool.Get().(*[]byte)
		defer bufferPool.Put(readMessageBuf)

		for {
			size, err := grpc.ReadGRPCMessage(r.Body, *readMessageBuf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			msg, err := unmarshalRequest(method.Input(), (*readMessageBuf)[:size])
			if err != nil {
				return err
			}
			reqFrameTracker.Add(msg)

			reqBytes, err := proto.Marshal(msg)
			if err != nil {
				msg.Release()
				return err
			}
			dynamicReq := dynamicpb.NewMessage(method.Input())
			if err := proto.Unmarshal(reqBytes, dynamicReq); err != nil {
				msg.Release()
				return err
			}

			if firstReq == nil {
				firstReq = msg
			} else {
				msg.Release()
			}

			if err := stream.Send(dynamicReq); err != nil {
				if firstReq != nil {
					firstReq.(releasableMessage).Release()
				}
				return err
			}
		}

		if firstReq != nil {
			firstReq.(releasableMessage).Release()
		}

		resp, err := stream.CloseAndReceive()
		if err != nil {
			return err
		}

		copyHeaders(resp.Header(), w.Header())
		w.Header().Set("x-fauxrpc-source", "proxy")
		*responseBody = resp.Msg

		respBytes, err := proto.Marshal(resp.Msg)
		if err != nil {
			return err
		}
		return grpc.WriteGRPCMessage(w, respBytes)
	}

	if isClientStream && isServerStream {
		bidiStream := client.CallBidiStream(ctx)
		copyHeaders(r.Header, bidiStream.RequestHeader())

		var firstReq proto.Message

		eg, _ := errgroup.WithContext(ctx)
		eg.Go(func() error {
			readMessageBuf := bufferPool.Get().(*[]byte)
			defer bufferPool.Put(readMessageBuf)

			for {
				size, err := grpc.ReadGRPCMessage(r.Body, *readMessageBuf)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
				msg, err := unmarshalRequest(method.Input(), (*readMessageBuf)[:size])
				if err != nil {
					return err
				}
				reqFrameTracker.Add(msg)

				reqBytes, err := proto.Marshal(msg)
				if err != nil {
					msg.Release()
					return err
				}
				dynamicReq := dynamicpb.NewMessage(method.Input())
				if err := proto.Unmarshal(reqBytes, dynamicReq); err != nil {
					msg.Release()
					return err
				}

				if firstReq == nil {
					firstReq = msg
				} else {
					msg.Release()
				}

				if err := bidiStream.Send(dynamicReq); err != nil {
					return err
				}
			}
			return bidiStream.CloseRequest()
		})

		eg.Go(func() error {
			for {
				respMsg, err := bidiStream.Receive()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				resFrameTracker.Add(respMsg)

				respBytes, err := proto.Marshal(respMsg)
				if err != nil {
					return err
				}
				if err := grpc.WriteGRPCMessage(w, respBytes); err != nil {
					return err
				}
			}
		})

		err := eg.Wait()
		if firstReq != nil {
			firstReq.(releasableMessage).Release()
		}

		copyHeaders(bidiStream.ResponseHeader(), w.Header())
		w.Header().Set("x-fauxrpc-source", "proxy")

		return err
	}

	return nil
}

func readUnaryRequest(r *http.Request, md protoreflect.MessageDescriptor) (releasableMessage, *status.Status) {
	readMessageBuf := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(readMessageBuf)

	size, err := grpc.ReadGRPCMessage(r.Body, *readMessageBuf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, status.New(codes.NotFound, err.Error())
	}
	msg, err := unmarshalRequest(md, (*readMessageBuf)[:size])
	if err != nil {
		return nil, status.New(codes.NotFound, err.Error())
	}
	return msg, nil
}
