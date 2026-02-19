package grpc

import (
	"io"
	"testing"
)

func BenchmarkWriteGRPCMessage(b *testing.B) {
	msg := []byte("hello world")
	w := io.Discard

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := WriteGRPCMessage(w, msg); err != nil {
			b.Fatal(err)
		}
	}
}
