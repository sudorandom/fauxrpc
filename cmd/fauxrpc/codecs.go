package main

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// dynamicProtoCodec implements connect.Codec to handle dynamicpb.Message responses.
type dynamicProtoCodec struct {
	methodDesc protoreflect.MethodDescriptor
}

func (c *dynamicProtoCodec) Name() string {
	return "proto"
}

func (c *dynamicProtoCodec) Marshal(msg any) ([]byte, error) {
	switch m := msg.(type) {
	case proto.Message:
		return proto.Marshal(m)

	case **dynamicpb.Message:
		if *m == nil {
			return nil, fmt.Errorf("cannot marshal nil **dynamicpb.Message")
		}
		return proto.Marshal(*m)

	default:
		return nil, fmt.Errorf("can't marshal %T", msg)
	}
}

// Unmarshal decodes a binary message into a dynamicpb.Message.
func (c *dynamicProtoCodec) Unmarshal(binary []byte, msg any) error {
	// Check if we're unmarshaling into a *dynamicpb.Message.
	// connect-go will pass a pointer to a nil *dynamicpb.Message.
	if ptr, ok := msg.(**dynamicpb.Message); ok {
		// Create a new dynamic message with the correct output descriptor.
		newMsg := dynamicpb.NewMessage(c.methodDesc.Output())
		// Unmarshal into the new message.
		if err := proto.Unmarshal(binary, newMsg); err != nil {
			return err
		}
		// Point the original pointer to the new message.
		*ptr = newMsg
		return nil
	}
	p, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("can't unmarshal into %T", msg)
	}
	return proto.Unmarshal(binary, p)
}

// dynamicProtoCodec implements connect.Codec to handle dynamicpb.Message responses.
type dynamicJSONCodec struct {
	methodDesc protoreflect.MethodDescriptor
}

func (c *dynamicJSONCodec) Name() string {
	return "json"
}

func (c *dynamicJSONCodec) Marshal(msg any) ([]byte, error) {
	switch m := msg.(type) {
	case proto.Message:
		return protojson.Marshal(m)

	case **dynamicpb.Message:
		if *m == nil {
			return nil, fmt.Errorf("cannot marshal nil **dynamicpb.Message")
		}
		return protojson.Marshal(*m)

	default:
		return nil, fmt.Errorf("can't marshal %T", msg)
	}
}

// Unmarshal decodes a JSON message into a dynamicpb.Message.
func (c *dynamicJSONCodec) Unmarshal(binary []byte, msg any) error {
	// Check if we're unmarshaling into a *dynamicpb.Message.
	// connect-go will pass a pointer to a nil *dynamicpb.Message.
	if ptr, ok := msg.(**dynamicpb.Message); ok {
		// Create a new dynamic message with the correct output descriptor.
		newMsg := dynamicpb.NewMessage(c.methodDesc.Output())
		// Unmarshal into the new message.
		if err := protojson.Unmarshal(binary, newMsg); err != nil {
			return err
		}
		// Point the original pointer to the new message.
		*ptr = newMsg
		return nil
	}
	p, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("can't unmarshal into %T", msg)
	}
	return protojson.Unmarshal(binary, p)
}
