package server

import (
	"net/http"
	"strings"
)

// responseCompressionEncoding returns the compression encoding the server
// should use for response frames, or "" for identity. Per the gRPC spec:
//
//   - grpc-accept-encoding lists the compression algorithms the client can
//     decode in responses (comma-separated, case-insensitive). Prefer gzip,
//     then deflate.
//
//   - If grpc-accept-encoding is absent but the client sent a compressed
//     request, it implicitly accepts responses using the same encoding,
//     so we mirror the encoding.
//
// Reference: https://grpc.github.io/grpc/core/md_doc_compression.html
func responseCompressionEncoding(r *http.Request) string {
	if accept := r.Header.Get("grpc-accept-encoding"); accept != "" {
		// grpc-accept-encoding is the authoritative list of what the client can
		// decode. Only use the grpc-encoding fallback when it is absent.
		var acceptsDeflate bool
		for enc := range strings.SplitSeq(accept, ",") {
			switch strings.ToLower(strings.TrimSpace(enc)) {
			case "gzip":
				return "gzip"
			case "deflate":
				acceptsDeflate = true
			}
		}
		if acceptsDeflate {
			return "deflate"
		}
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("grpc-encoding"))) {
	case "gzip":
		return "gzip"
	case "deflate":
		return "deflate"
	default:
		return ""
	}
}

// clientAcceptsGzip reports whether the client has indicated it can decode
// gzip-compressed response frames.
func clientAcceptsGzip(r *http.Request) bool {
	return responseCompressionEncoding(r) == "gzip"
}
