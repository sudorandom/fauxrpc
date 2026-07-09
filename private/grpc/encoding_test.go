package grpc

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeUncompressedFrame builds a raw gRPC length-prefixed frame with flag=0.
func makeUncompressedFrame(payload []byte) []byte {
	frame := make([]byte, 5+len(payload))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:], uint32(len(payload)))
	copy(frame[5:], payload)
	return frame
}

// makeGzipFrame builds a raw gRPC length-prefixed frame with flag=1 (gzip).
func makeGzipFrame(payload []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(payload)
	_ = gz.Close()
	compressed := buf.Bytes()

	frame := make([]byte, 5+len(compressed))
	frame[0] = 1
	binary.BigEndian.PutUint32(frame[1:], uint32(len(compressed)))
	copy(frame[5:], compressed)
	return frame
}

// --- ReadGRPCMessage ---

func TestReadGRPCMessage_Uncompressed(t *testing.T) {
	payload := []byte("hello world")
	frame := makeUncompressedFrame(payload)

	buf := make([]byte, 4096)
	n, err := ReadGRPCMessage(bytes.NewReader(frame), buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

func TestReadGRPCMessage_Gzip(t *testing.T) {
	payload := []byte("hello compressed world")
	frame := makeGzipFrame(payload)

	buf := make([]byte, 4096)
	n, err := ReadGRPCMessage(bytes.NewReader(frame), buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

func TestReadGRPCMessage_EmptyPayload(t *testing.T) {
	frame := makeUncompressedFrame([]byte{})

	buf := make([]byte, 4096)
	n, err := ReadGRPCMessage(bytes.NewReader(frame), buf)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestReadGRPCMessage_EOF(t *testing.T) {
	_, err := ReadGRPCMessage(bytes.NewReader([]byte{}), make([]byte, 4096))
	assert.ErrorIs(t, err, io.EOF)
}

func TestReadGRPCMessage_MultipleFrames(t *testing.T) {
	// Write two frames back-to-back and read them sequentially.
	payload1 := []byte("frame one")
	payload2 := []byte("frame two")

	var buf bytes.Buffer
	buf.Write(makeUncompressedFrame(payload1))
	buf.Write(makeGzipFrame(payload2))

	readBuf := make([]byte, 4096)
	r := bytes.NewReader(buf.Bytes())

	n1, err := ReadGRPCMessage(r, readBuf)
	require.NoError(t, err)
	assert.Equal(t, payload1, readBuf[:n1])

	n2, err := ReadGRPCMessage(r, readBuf)
	require.NoError(t, err)
	assert.Equal(t, payload2, readBuf[:n2])
}

func TestReadGRPCMessage_InvalidGzip(t *testing.T) {
	// Flag=1 but payload is not valid gzip.
	bad := []byte("this is not gzip data")
	frame := make([]byte, 5+len(bad))
	frame[0] = 1
	binary.BigEndian.PutUint32(frame[1:], uint32(len(bad)))
	copy(frame[5:], bad)

	buf := make([]byte, 4096)
	_, err := ReadGRPCMessage(bytes.NewReader(frame), buf)
	require.Error(t, err)
}

// --- WriteGRPCMessage ---

func TestWriteGRPCMessage_SetsUncompressedFlag(t *testing.T) {
	payload := []byte("hello")
	var buf bytes.Buffer
	require.NoError(t, WriteGRPCMessage(&buf, payload))

	b := buf.Bytes()
	require.Len(t, b, 5+len(payload))
	assert.Equal(t, byte(0), b[0], "compression flag should be 0")
	size := binary.BigEndian.Uint32(b[1:5])
	assert.Equal(t, uint32(len(payload)), size)
	assert.Equal(t, payload, b[5:])
}

// --- WriteGRPCMessageGzip ---

func TestWriteGRPCMessageGzip_SetsCompressedFlag(t *testing.T) {
	payload := []byte("hello compressed")
	var buf bytes.Buffer
	require.NoError(t, WriteGRPCMessageGzip(&buf, payload))

	b := buf.Bytes()
	require.GreaterOrEqual(t, len(b), 5)
	assert.Equal(t, byte(1), b[0], "compression flag should be 1")

	size := binary.BigEndian.Uint32(b[1:5])
	compressed := b[5 : 5+size]
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)
	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, payload, decompressed)
}

// --- Round-trip ---

func TestRoundTrip_WriteGzip_ReadGzip(t *testing.T) {
	payload := []byte("round-trip test payload, larger than typical gzip overhead to actually compress well")
	var buf bytes.Buffer

	require.NoError(t, WriteGRPCMessageGzip(&buf, payload))

	readBuf := make([]byte, 4096)
	n, err := ReadGRPCMessage(&buf, readBuf)
	require.NoError(t, err)
	assert.Equal(t, payload, readBuf[:n])
}

func TestRoundTrip_WriteUncompressed_ReadUncompressed(t *testing.T) {
	payload := []byte("uncompressed round-trip")
	var buf bytes.Buffer

	require.NoError(t, WriteGRPCMessage(&buf, payload))

	readBuf := make([]byte, 4096)
	n, err := ReadGRPCMessage(&buf, readBuf)
	require.NoError(t, err)
	assert.Equal(t, payload, readBuf[:n])
}

// --- Benchmarks ---

func BenchmarkWriteGRPCMessageGzip(b *testing.B) {
	msg := []byte("hello world — this is a benchmark payload for gzip compression performance")
	w := io.Discard

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := WriteGRPCMessageGzip(w, msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadGRPCMessage_Gzip(b *testing.B) {
	payload := []byte("hello world — this is a benchmark payload for gzip decompression performance")
	frame := makeGzipFrame(payload)
	buf := make([]byte, 4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ReadGRPCMessage(bytes.NewReader(frame), buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
