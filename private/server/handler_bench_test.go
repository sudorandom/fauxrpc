package server

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	fauxlog "github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func BenchmarkHandler_RequestAllocation(b *testing.B) {
	// Setup
	logger := fauxlog.NewLogger()
	s := &mockServer{
		ServiceRegistry: mustNewRegistry(),
		StubDatabase:    stubs.NewStubDatabase(),
		logger:          logger,
	}

	validator, err := protovalidate.New()
	require.NoError(b, err)

	faker := fauxrpc.NewFauxFaker()

	// Eliza Service
	file := elizav1.File_connectrpc_eliza_v1_eliza_proto
	service := file.Services().ByName("ElizaService")
	require.NotNil(b, service)

	handler := NewHandler(service, faker, validator, s, logger)

	// Use Converse method
	method := service.Methods().ByName("Converse")
	require.NotNil(b, method)
	url := "/connectrpc.eliza.v1.ElizaService/Converse"

	// Prepare a message
	msg := &elizav1.ConverseRequest{Sentence: "Hello World"}
	msgBytes, err := proto.Marshal(msg)
	require.NoError(b, err)

	// Prepare framed message
	framedMsg := make([]byte, 5+len(msgBytes))
	framedMsg[0] = 0 // not compressed
	length := len(msgBytes)
	binary.BigEndian.PutUint32(framedMsg[1:], uint32(length))
	copy(framedMsg[5:], msgBytes)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", url, bytes.NewReader(framedMsg))
		req.Header.Set("Content-Type", "application/grpc")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
	}
}

type mockWriter struct {
	h http.Header
}

func (m *mockWriter) Header() http.Header {
	return m.h
}

func (m *mockWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockWriter) WriteHeader(statusCode int) {}

func BenchmarkGRPCWriteStatus(b *testing.B) {
	st := status.New(codes.NotFound, "not found")
	w := &mockWriter{h: make(http.Header)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		grpcWriteStatus(w, st)
		// Reset header
		for k := range w.h {
			delete(w.h, k)
		}
	}
}
