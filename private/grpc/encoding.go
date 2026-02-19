package grpc

import (
	"encoding/binary"
	"fmt"
	"io"
)

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

func ReadGRPCMessage(body io.Reader, msg []byte) (int, error) {
	prefixes := [5]byte{}
	if _, err := io.ReadFull(body, prefixes[:]); err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, fmt.Errorf("failed to read envelope: %w", err)
	}

	msgSize := int64(binary.BigEndian.Uint32(prefixes[1:5]))
	if msgSize == 0 {
		return 0, nil
	}
	return body.Read(msg[:msgSize])
}
