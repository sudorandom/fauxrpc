package protobuf

import (
	"encoding/binary"
	"io"
)

func writeGRPCMessage(w io.Writer, msg []byte) {
	prefix := make([]byte, 5)
	binary.BigEndian.PutUint32(prefix[1:], uint32(len(msg)))
	w.Write(prefix)
	w.Write(msg)
}
