package grpc

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

var gzipReaderPool = sync.Pool{
	New: func() any { return new(gzip.Reader) },
}

var gzipWriterPool = sync.Pool{
	New: func() any { return gzip.NewWriter(io.Discard) },
}

// WriteGRPCMessage writes an uncompressed gRPC length-prefixed message.
func WriteGRPCMessage(w io.Writer, msg []byte) error {
	var prefix [5]byte
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(msg)))
	if _, err := w.Write(prefix[:]); err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return nil
}

// WriteGRPCMessageGzip writes a gzip-compressed gRPC length-prefixed message.
// The compression flag byte is set to 1.
func WriteGRPCMessageGzip(w io.Writer, msg []byte) error {
	var buf bytes.Buffer
	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(&buf)
	if _, err := gz.Write(msg); err != nil {
		gzipWriterPool.Put(gz)
		return fmt.Errorf("grpc: gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		gzipWriterPool.Put(gz)
		return fmt.Errorf("grpc: gzip close: %w", err)
	}
	gzipWriterPool.Put(gz)

	compressed := buf.Bytes()
	var prefix [5]byte
	prefix[0] = 1 // compressed
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(compressed)))
	if _, err := w.Write(prefix[:]); err != nil {
		return err
	}
	if _, err := w.Write(compressed); err != nil {
		return err
	}
	return nil
}

// ReadGRPCMessage reads a gRPC length-prefixed message from body into msg.
// If the compression flag is set to 1 (gzip), the payload is decompressed
// transparently before being written into msg.
func ReadGRPCMessage(body io.Reader, msg []byte) (int, error) {
	prefixes := [5]byte{}
	if _, err := io.ReadFull(body, prefixes[:]); err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, fmt.Errorf("failed to read envelope: %w", err)
	}

	isCompressed := prefixes[0] == 1
	msgSize := int64(binary.BigEndian.Uint32(prefixes[1:5]))
	if msgSize == 0 {
		return 0, nil
	}

	n, err := io.ReadFull(body, msg[:msgSize])
	if err != nil {
		return n, fmt.Errorf("failed to read message body: %w", err)
	}

	if !isCompressed {
		return n, nil
	}

	// Decompress gzip-encoded payload.
	gr := gzipReaderPool.Get().(*gzip.Reader)
	if resetErr := gr.Reset(bytes.NewReader(msg[:n])); resetErr != nil {
		gzipReaderPool.Put(gr)
		return 0, fmt.Errorf("failed to init gzip reader: %w", resetErr)
	}
	decompressed, readErr := io.ReadAll(gr)
	closeErr := gr.Close()
	gzipReaderPool.Put(gr)
	if readErr != nil {
		return 0, fmt.Errorf("failed to decompress message: %w", readErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("failed to close gzip reader: %w", closeErr)
	}
	copy(msg, decompressed)
	return len(decompressed), nil
}
