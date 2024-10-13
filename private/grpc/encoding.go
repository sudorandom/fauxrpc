package grpc

import (
	"encoding/binary"
	"io"
)

func WriteGRPCMessage(w io.Writer, msg []byte) {
	prefix := make([]byte, 5)
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(msg)))
	_, _ = w.Write(prefix)
	_, _ = w.Write(msg)
}
