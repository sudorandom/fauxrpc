package grpc

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrMissingCompressionEncoding = errors.New("compressed message missing grpc-encoding")

type UnsupportedCompressionError struct {
	Encoding string
}

func (e *UnsupportedCompressionError) Error() string {
	return fmt.Sprintf("unsupported grpc-encoding %q", e.Encoding)
}

type compressionCodec interface {
	NewReader(io.Reader) (io.ReadCloser, error)
	NewWriter(io.Writer) (io.WriteCloser, error)
}

type compressionCodecFuncs struct {
	newReader func(io.Reader) (io.ReadCloser, error)
	newWriter func(io.Writer) (io.WriteCloser, error)
}

func (c compressionCodecFuncs) NewReader(r io.Reader) (io.ReadCloser, error) {
	return c.newReader(r)
}

func (c compressionCodecFuncs) NewWriter(w io.Writer) (io.WriteCloser, error) {
	return c.newWriter(w)
}

var compressionCodecs = map[string]compressionCodec{
	"gzip": compressionCodecFuncs{
		newReader: func(r io.Reader) (io.ReadCloser, error) {
			return gzip.NewReader(r)
		},
		newWriter: func(w io.Writer) (io.WriteCloser, error) {
			return gzip.NewWriter(w), nil
		},
	},
	"deflate": compressionCodecFuncs{
		newReader: func(r io.Reader) (io.ReadCloser, error) {
			return zlib.NewReader(r)
		},
		newWriter: func(w io.Writer) (io.WriteCloser, error) {
			return zlib.NewWriter(w), nil
		},
	},
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
	return WriteGRPCMessageCompressed(w, msg, "gzip")
}

// WriteGRPCMessageCompressed writes a compressed gRPC length-prefixed message.
// The compression flag byte is set to 1.
func WriteGRPCMessageCompressed(w io.Writer, msg []byte, encoding string) error {
	codec, ok := lookupCompressionCodec(encoding)
	if !ok {
		return fmt.Errorf("grpc: unsupported compression encoding %q", encoding)
	}

	var buf bytes.Buffer
	if err := writeCompressedPayload(&buf, msg, codec); err != nil {
		return fmt.Errorf("grpc: %s write: %w", encoding, err)
	}

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

func lookupCompressionCodec(encoding string) (compressionCodec, bool) {
	codec, ok := compressionCodecs[strings.ToLower(strings.TrimSpace(encoding))]
	return codec, ok
}

func writeCompressedPayload(w io.Writer, msg []byte, codec compressionCodec) error {
	cw, err := codec.NewWriter(w)
	if err != nil {
		return err
	}
	if _, err := cw.Write(msg); err != nil {
		_ = cw.Close()
		return err
	}
	return cw.Close()
}

// ReadGRPCMessage reads a gRPC length-prefixed message from body into msg.
// If the compression flag is set to 1 (gzip), the payload is decompressed
// transparently before being written into msg.
func ReadGRPCMessage(body io.Reader, msg []byte) (int, error) {
	return ReadGRPCMessageWithEncoding(body, msg, "gzip")
}

// ReadGRPCMessageWithEncoding reads a gRPC length-prefixed message from body
// into msg. If the compression flag is set to 1, the payload is decompressed
// using encoding. The gRPC "deflate" encoding is the zlib structure from RFC
// 1950 carrying the deflate algorithm from RFC 1951, never raw deflate data.
func ReadGRPCMessageWithEncoding(body io.Reader, msg []byte, encoding string) (int, error) {
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

	if strings.TrimSpace(encoding) == "" || strings.EqualFold(strings.TrimSpace(encoding), "identity") {
		return 0, ErrMissingCompressionEncoding
	}
	codec, ok := lookupCompressionCodec(encoding)
	if !ok {
		return 0, &UnsupportedCompressionError{Encoding: strings.TrimSpace(encoding)}
	}
	return readCompressedPayload(msg, n, codec)
}

func readCompressedPayload(msg []byte, n int, codec compressionCodec) (int, error) {
	cr, err := codec.NewReader(bytes.NewReader(msg[:n]))
	if err != nil {
		return 0, fmt.Errorf("failed to init compression reader: %w", err)
	}
	decompressed, readErr := io.ReadAll(cr)
	closeErr := cr.Close()
	if readErr != nil {
		return 0, fmt.Errorf("failed to decompress message: %w", readErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("failed to close compression reader: %w", closeErr)
	}
	copy(msg, decompressed)
	return len(decompressed), nil
}
