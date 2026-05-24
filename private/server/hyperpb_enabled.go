//go:build amd64 || arm64

package server

import (
	"sync"

	"buf.build/go/hyperpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	hyperpbTypes      sync.Map // map[protoreflect.FullName]*hyperpb.MessageType
	hyperpbSharedPool = sync.Pool{
		New: func() any {
			return new(hyperpb.Shared)
		},
	}
)

type releasableMessage interface {
	proto.Message
	Release()
}

type hyperpbMessage struct {
	*hyperpb.Message
	shared *hyperPBShared
}

func (m *hyperpbMessage) Release() {
	m.shared.Release()
}

type hyperPBShared struct {
	*hyperpb.Shared
}

func newHyperPBShared() *hyperPBShared {
	return &hyperPBShared{hyperpbSharedPool.Get().(*hyperpb.Shared)}
}

func (s *hyperPBShared) Release() {
	if s.Shared != nil {
		s.Free()
		hyperpbSharedPool.Put(s.Shared)
		s.Shared = nil
	}
}

func unmarshalRequest(md protoreflect.MessageDescriptor, data []byte) (releasableMessage, error) {
	var msgType *hyperpb.MessageType
	if v, ok := hyperpbTypes.Load(md.FullName()); ok {
		msgType = v.(*hyperpb.MessageType)
	} else {
		msgType = hyperpb.CompileMessageDescriptor(md)
		hyperpbTypes.Store(md.FullName(), msgType)
	}

	shared := newHyperPBShared()
	msg := shared.NewMessage(msgType)
	if err := msg.Unmarshal(data); err != nil {
		shared.Release()
		return nil, err
	}
	return &hyperpbMessage{Message: msg, shared: shared}, nil
}
