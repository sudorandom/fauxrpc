//go:build !amd64 && !arm64

package server

import (
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type releasableMessage interface {
	proto.Message
	Release()
}

type dynamicMessage struct {
	proto.Message
}

func (m *dynamicMessage) Release() {}

func unmarshalRequest(md protoreflect.MessageDescriptor, data []byte) (releasableMessage, error) {
	msg := registry.NewMessage(md).Interface()
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return &dynamicMessage{Message: msg}, nil
}
