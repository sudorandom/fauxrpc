// Copyright 2022-2023 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: connectrpc/eliza/v1/eliza.proto

package elizav1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SayRequest is a single-sentence request.
type SayRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sentence string `protobuf:"bytes,1,opt,name=sentence,proto3" json:"sentence,omitempty"`
}

func (x *SayRequest) Reset() {
	*x = SayRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SayRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SayRequest) ProtoMessage() {}

func (x *SayRequest) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SayRequest.ProtoReflect.Descriptor instead.
func (*SayRequest) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{0}
}

func (x *SayRequest) GetSentence() string {
	if x != nil {
		return x.Sentence
	}
	return ""
}

// SayResponse is a single-sentence response.
type SayResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sentence string `protobuf:"bytes,1,opt,name=sentence,proto3" json:"sentence,omitempty"`
}

func (x *SayResponse) Reset() {
	*x = SayResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SayResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SayResponse) ProtoMessage() {}

func (x *SayResponse) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SayResponse.ProtoReflect.Descriptor instead.
func (*SayResponse) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{1}
}

func (x *SayResponse) GetSentence() string {
	if x != nil {
		return x.Sentence
	}
	return ""
}

// ConverseRequest is a single sentence request sent as part of a
// back-and-forth conversation.
type ConverseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sentence string `protobuf:"bytes,1,opt,name=sentence,proto3" json:"sentence,omitempty"`
}

func (x *ConverseRequest) Reset() {
	*x = ConverseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConverseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConverseRequest) ProtoMessage() {}

func (x *ConverseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConverseRequest.ProtoReflect.Descriptor instead.
func (*ConverseRequest) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{2}
}

func (x *ConverseRequest) GetSentence() string {
	if x != nil {
		return x.Sentence
	}
	return ""
}

// ConverseResponse is a single sentence response sent in answer to a
// ConverseRequest.
type ConverseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sentence string `protobuf:"bytes,1,opt,name=sentence,proto3" json:"sentence,omitempty"`
}

func (x *ConverseResponse) Reset() {
	*x = ConverseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConverseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConverseResponse) ProtoMessage() {}

func (x *ConverseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConverseResponse.ProtoReflect.Descriptor instead.
func (*ConverseResponse) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{3}
}

func (x *ConverseResponse) GetSentence() string {
	if x != nil {
		return x.Sentence
	}
	return ""
}

// IntroduceRequest asks Eliza to introduce itself to the named user.
type IntroduceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *IntroduceRequest) Reset() {
	*x = IntroduceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IntroduceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IntroduceRequest) ProtoMessage() {}

func (x *IntroduceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IntroduceRequest.ProtoReflect.Descriptor instead.
func (*IntroduceRequest) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{4}
}

func (x *IntroduceRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// IntroduceResponse is one sentence of Eliza's introductory monologue.
type IntroduceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sentence string `protobuf:"bytes,1,opt,name=sentence,proto3" json:"sentence,omitempty"`
}

func (x *IntroduceResponse) Reset() {
	*x = IntroduceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IntroduceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IntroduceResponse) ProtoMessage() {}

func (x *IntroduceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_eliza_v1_eliza_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IntroduceResponse.ProtoReflect.Descriptor instead.
func (*IntroduceResponse) Descriptor() ([]byte, []int) {
	return file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP(), []int{5}
}

func (x *IntroduceResponse) GetSentence() string {
	if x != nil {
		return x.Sentence
	}
	return ""
}

var File_connectrpc_eliza_v1_eliza_proto protoreflect.FileDescriptor

var file_connectrpc_eliza_v1_eliza_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x6c, 0x69,
	0x7a, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c,
	0x69, 0x7a, 0x61, 0x2e, 0x76, 0x31, 0x22, 0x28, 0x0a, 0x0a, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65,
	0x22, 0x29, 0x0a, 0x0b, 0x53, 0x61, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x2d, 0x0a, 0x0f, 0x43,
	0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a,
	0x0a, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x2e, 0x0a, 0x10, 0x43, 0x6f,
	0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a,
	0x0a, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x26, 0x0a, 0x10, 0x49, 0x6e,
	0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x22, 0x2f, 0x0a, 0x11, 0x49, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65,
	0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x74, 0x65,
	0x6e, 0x63, 0x65, 0x32, 0x9c, 0x02, 0x0a, 0x0c, 0x45, 0x6c, 0x69, 0x7a, 0x61, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x4d, 0x0a, 0x03, 0x53, 0x61, 0x79, 0x12, 0x1f, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03,
	0x90, 0x02, 0x01, 0x12, 0x5d, 0x0a, 0x08, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x12,
	0x24, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69,
	0x7a, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72,
	0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x76,
	0x65, 0x72, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01,
	0x30, 0x01, 0x12, 0x5e, 0x0a, 0x09, 0x49, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x12,
	0x25, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69,
	0x7a, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x30, 0x01, 0x42, 0xf0, 0x01, 0x0a, 0x17, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2e, 0x76, 0x31, 0x42, 0x0a,
	0x45, 0x6c, 0x69, 0x7a, 0x61, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x5b, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x64, 0x6f, 0x72, 0x61, 0x6e,
	0x64, 0x6f, 0x6d, 0x2f, 0x66, 0x61, 0x75, 0x78, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2d, 0x66, 0x61,
	0x75, 0x78, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x2f, 0x76,
	0x31, 0x3b, 0x65, 0x6c, 0x69, 0x7a, 0x61, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x45, 0x58, 0xaa,
	0x02, 0x13, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6c, 0x69,
	0x7a, 0x61, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x13, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72,
	0x70, 0x63, 0x5c, 0x45, 0x6c, 0x69, 0x7a, 0x61, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1f, 0x43, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x5c, 0x45, 0x6c, 0x69, 0x7a, 0x61, 0x5c, 0x56,
	0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x15,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x3a, 0x3a, 0x45, 0x6c, 0x69, 0x7a,
	0x61, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_connectrpc_eliza_v1_eliza_proto_rawDescOnce sync.Once
	file_connectrpc_eliza_v1_eliza_proto_rawDescData = file_connectrpc_eliza_v1_eliza_proto_rawDesc
)

func file_connectrpc_eliza_v1_eliza_proto_rawDescGZIP() []byte {
	file_connectrpc_eliza_v1_eliza_proto_rawDescOnce.Do(func() {
		file_connectrpc_eliza_v1_eliza_proto_rawDescData = protoimpl.X.CompressGZIP(file_connectrpc_eliza_v1_eliza_proto_rawDescData)
	})
	return file_connectrpc_eliza_v1_eliza_proto_rawDescData
}

var file_connectrpc_eliza_v1_eliza_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_connectrpc_eliza_v1_eliza_proto_goTypes = []any{
	(*SayRequest)(nil),        // 0: connectrpc.eliza.v1.SayRequest
	(*SayResponse)(nil),       // 1: connectrpc.eliza.v1.SayResponse
	(*ConverseRequest)(nil),   // 2: connectrpc.eliza.v1.ConverseRequest
	(*ConverseResponse)(nil),  // 3: connectrpc.eliza.v1.ConverseResponse
	(*IntroduceRequest)(nil),  // 4: connectrpc.eliza.v1.IntroduceRequest
	(*IntroduceResponse)(nil), // 5: connectrpc.eliza.v1.IntroduceResponse
}
var file_connectrpc_eliza_v1_eliza_proto_depIdxs = []int32{
	0, // 0: connectrpc.eliza.v1.ElizaService.Say:input_type -> connectrpc.eliza.v1.SayRequest
	2, // 1: connectrpc.eliza.v1.ElizaService.Converse:input_type -> connectrpc.eliza.v1.ConverseRequest
	4, // 2: connectrpc.eliza.v1.ElizaService.Introduce:input_type -> connectrpc.eliza.v1.IntroduceRequest
	1, // 3: connectrpc.eliza.v1.ElizaService.Say:output_type -> connectrpc.eliza.v1.SayResponse
	3, // 4: connectrpc.eliza.v1.ElizaService.Converse:output_type -> connectrpc.eliza.v1.ConverseResponse
	5, // 5: connectrpc.eliza.v1.ElizaService.Introduce:output_type -> connectrpc.eliza.v1.IntroduceResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_connectrpc_eliza_v1_eliza_proto_init() }
func file_connectrpc_eliza_v1_eliza_proto_init() {
	if File_connectrpc_eliza_v1_eliza_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SayRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SayResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ConverseRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*ConverseResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*IntroduceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_eliza_v1_eliza_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*IntroduceResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connectrpc_eliza_v1_eliza_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_connectrpc_eliza_v1_eliza_proto_goTypes,
		DependencyIndexes: file_connectrpc_eliza_v1_eliza_proto_depIdxs,
		MessageInfos:      file_connectrpc_eliza_v1_eliza_proto_msgTypes,
	}.Build()
	File_connectrpc_eliza_v1_eliza_proto = out.File
	file_connectrpc_eliza_v1_eliza_proto_rawDesc = nil
	file_connectrpc_eliza_v1_eliza_proto_goTypes = nil
	file_connectrpc_eliza_v1_eliza_proto_depIdxs = nil
}
