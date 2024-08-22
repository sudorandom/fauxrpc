// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: stubs/v1/stubs_service.proto

package stubsv1

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

type AddStubsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stubs []*Stub `protobuf:"bytes,1,rep,name=stubs,proto3" json:"stubs,omitempty"`
}

func (x *AddStubsRequest) Reset() {
	*x = AddStubsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddStubsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddStubsRequest) ProtoMessage() {}

func (x *AddStubsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddStubsRequest.ProtoReflect.Descriptor instead.
func (*AddStubsRequest) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{0}
}

func (x *AddStubsRequest) GetStubs() []*Stub {
	if x != nil {
		return x.Stubs
	}
	return nil
}

type AddStubsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stubs []*Stub `protobuf:"bytes,1,rep,name=stubs,proto3" json:"stubs,omitempty"`
}

func (x *AddStubsResponse) Reset() {
	*x = AddStubsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddStubsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddStubsResponse) ProtoMessage() {}

func (x *AddStubsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddStubsResponse.ProtoReflect.Descriptor instead.
func (*AddStubsResponse) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{1}
}

func (x *AddStubsResponse) GetStubs() []*Stub {
	if x != nil {
		return x.Stubs
	}
	return nil
}

type RemoveStubsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StubRefs []*StubRef `protobuf:"bytes,1,rep,name=stub_refs,json=stubRefs,proto3" json:"stub_refs,omitempty"`
}

func (x *RemoveStubsRequest) Reset() {
	*x = RemoveStubsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveStubsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveStubsRequest) ProtoMessage() {}

func (x *RemoveStubsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveStubsRequest.ProtoReflect.Descriptor instead.
func (*RemoveStubsRequest) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{2}
}

func (x *RemoveStubsRequest) GetStubRefs() []*StubRef {
	if x != nil {
		return x.StubRefs
	}
	return nil
}

type RemoveStubsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveStubsResponse) Reset() {
	*x = RemoveStubsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveStubsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveStubsResponse) ProtoMessage() {}

func (x *RemoveStubsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveStubsResponse.ProtoReflect.Descriptor instead.
func (*RemoveStubsResponse) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{3}
}

type RemoveAllStubsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveAllStubsRequest) Reset() {
	*x = RemoveAllStubsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveAllStubsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveAllStubsRequest) ProtoMessage() {}

func (x *RemoveAllStubsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveAllStubsRequest.ProtoReflect.Descriptor instead.
func (*RemoveAllStubsRequest) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{4}
}

type RemoveAllStubsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveAllStubsResponse) Reset() {
	*x = RemoveAllStubsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveAllStubsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveAllStubsResponse) ProtoMessage() {}

func (x *RemoveAllStubsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveAllStubsResponse.ProtoReflect.Descriptor instead.
func (*RemoveAllStubsResponse) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{5}
}

type ListStubsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StubRef *StubRef `protobuf:"bytes,1,opt,name=stub_ref,json=stubRef,proto3" json:"stub_ref,omitempty"`
}

func (x *ListStubsRequest) Reset() {
	*x = ListStubsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListStubsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListStubsRequest) ProtoMessage() {}

func (x *ListStubsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListStubsRequest.ProtoReflect.Descriptor instead.
func (*ListStubsRequest) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{6}
}

func (x *ListStubsRequest) GetStubRef() *StubRef {
	if x != nil {
		return x.StubRef
	}
	return nil
}

type ListStubsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stubs []*Stub `protobuf:"bytes,1,rep,name=stubs,proto3" json:"stubs,omitempty"`
}

func (x *ListStubsResponse) Reset() {
	*x = ListStubsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListStubsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListStubsResponse) ProtoMessage() {}

func (x *ListStubsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListStubsResponse.ProtoReflect.Descriptor instead.
func (*ListStubsResponse) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_service_proto_rawDescGZIP(), []int{7}
}

func (x *ListStubsResponse) GetStubs() []*Stub {
	if x != nil {
		return x.Stubs
	}
	return nil
}

var File_stubs_v1_stubs_service_proto protoreflect.FileDescriptor

var file_stubs_v1_stubs_service_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x73,
	0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08,
	0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x14, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2f,
	0x76, 0x31, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x37,
	0x0a, 0x0f, 0x41, 0x64, 0x64, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x24, 0x0a, 0x05, 0x73, 0x74, 0x75, 0x62, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0e, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x75, 0x62,
	0x52, 0x05, 0x73, 0x74, 0x75, 0x62, 0x73, 0x22, 0x38, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x53, 0x74,
	0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x73,
	0x74, 0x75, 0x62, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x75,
	0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x75, 0x62, 0x52, 0x05, 0x73, 0x74, 0x75, 0x62,
	0x73, 0x22, 0x44, 0x0a, 0x12, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x53, 0x74, 0x75, 0x62, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x09, 0x73, 0x74, 0x75, 0x62, 0x5f,
	0x72, 0x65, 0x66, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x75,
	0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x75, 0x62, 0x52, 0x65, 0x66, 0x52, 0x08, 0x73,
	0x74, 0x75, 0x62, 0x52, 0x65, 0x66, 0x73, 0x22, 0x15, 0x0a, 0x13, 0x52, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x17,
	0x0a, 0x15, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x41, 0x6c, 0x6c, 0x53, 0x74, 0x75, 0x62, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x18, 0x0a, 0x16, 0x52, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x41, 0x6c, 0x6c, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x40, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x08, 0x73, 0x74, 0x75, 0x62, 0x5f, 0x72, 0x65,
	0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x74, 0x75, 0x62, 0x52, 0x65, 0x66, 0x52, 0x07, 0x73, 0x74, 0x75, 0x62,
	0x52, 0x65, 0x66, 0x22, 0x39, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x75, 0x62, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x73, 0x74, 0x75, 0x62,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x74, 0x75, 0x62, 0x52, 0x05, 0x73, 0x74, 0x75, 0x62, 0x73, 0x32, 0xc0,
	0x02, 0x0a, 0x0c, 0x53, 0x74, 0x75, 0x62, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x43, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x53, 0x74, 0x75, 0x62, 0x73, 0x12, 0x19, 0x2e, 0x73, 0x74,
	0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x4c, 0x0a, 0x0b, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x53, 0x74,
	0x75, 0x62, 0x73, 0x12, 0x1c, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x55, 0x0a, 0x0e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x41, 0x6c, 0x6c, 0x53,
	0x74, 0x75, 0x62, 0x73, 0x12, 0x1f, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x41, 0x6c, 0x6c, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x41, 0x6c, 0x6c, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x46, 0x0a, 0x09, 0x4c, 0x69, 0x73,
	0x74, 0x53, 0x74, 0x75, 0x62, 0x73, 0x12, 0x1a, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76,
	0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x53, 0x74, 0x75, 0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0xa4, 0x01, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e,
	0x76, 0x31, 0x42, 0x11, 0x53, 0x74, 0x75, 0x62, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x64, 0x6f, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2f, 0x66,
	0x61, 0x75, 0x78, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2f, 0x76,
	0x31, 0x3b, 0x73, 0x74, 0x75, 0x62, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa,
	0x02, 0x08, 0x53, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x08, 0x53, 0x74, 0x75,
	0x62, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x53, 0x74, 0x75, 0x62, 0x73, 0x5c, 0x56, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x53,
	0x74, 0x75, 0x62, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stubs_v1_stubs_service_proto_rawDescOnce sync.Once
	file_stubs_v1_stubs_service_proto_rawDescData = file_stubs_v1_stubs_service_proto_rawDesc
)

func file_stubs_v1_stubs_service_proto_rawDescGZIP() []byte {
	file_stubs_v1_stubs_service_proto_rawDescOnce.Do(func() {
		file_stubs_v1_stubs_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_stubs_v1_stubs_service_proto_rawDescData)
	})
	return file_stubs_v1_stubs_service_proto_rawDescData
}

var file_stubs_v1_stubs_service_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_stubs_v1_stubs_service_proto_goTypes = []any{
	(*AddStubsRequest)(nil),        // 0: stubs.v1.AddStubsRequest
	(*AddStubsResponse)(nil),       // 1: stubs.v1.AddStubsResponse
	(*RemoveStubsRequest)(nil),     // 2: stubs.v1.RemoveStubsRequest
	(*RemoveStubsResponse)(nil),    // 3: stubs.v1.RemoveStubsResponse
	(*RemoveAllStubsRequest)(nil),  // 4: stubs.v1.RemoveAllStubsRequest
	(*RemoveAllStubsResponse)(nil), // 5: stubs.v1.RemoveAllStubsResponse
	(*ListStubsRequest)(nil),       // 6: stubs.v1.ListStubsRequest
	(*ListStubsResponse)(nil),      // 7: stubs.v1.ListStubsResponse
	(*Stub)(nil),                   // 8: stubs.v1.Stub
	(*StubRef)(nil),                // 9: stubs.v1.StubRef
}
var file_stubs_v1_stubs_service_proto_depIdxs = []int32{
	8, // 0: stubs.v1.AddStubsRequest.stubs:type_name -> stubs.v1.Stub
	8, // 1: stubs.v1.AddStubsResponse.stubs:type_name -> stubs.v1.Stub
	9, // 2: stubs.v1.RemoveStubsRequest.stub_refs:type_name -> stubs.v1.StubRef
	9, // 3: stubs.v1.ListStubsRequest.stub_ref:type_name -> stubs.v1.StubRef
	8, // 4: stubs.v1.ListStubsResponse.stubs:type_name -> stubs.v1.Stub
	0, // 5: stubs.v1.StubsService.AddStubs:input_type -> stubs.v1.AddStubsRequest
	2, // 6: stubs.v1.StubsService.RemoveStubs:input_type -> stubs.v1.RemoveStubsRequest
	4, // 7: stubs.v1.StubsService.RemoveAllStubs:input_type -> stubs.v1.RemoveAllStubsRequest
	6, // 8: stubs.v1.StubsService.ListStubs:input_type -> stubs.v1.ListStubsRequest
	1, // 9: stubs.v1.StubsService.AddStubs:output_type -> stubs.v1.AddStubsResponse
	3, // 10: stubs.v1.StubsService.RemoveStubs:output_type -> stubs.v1.RemoveStubsResponse
	5, // 11: stubs.v1.StubsService.RemoveAllStubs:output_type -> stubs.v1.RemoveAllStubsResponse
	7, // 12: stubs.v1.StubsService.ListStubs:output_type -> stubs.v1.ListStubsResponse
	9, // [9:13] is the sub-list for method output_type
	5, // [5:9] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_stubs_v1_stubs_service_proto_init() }
func file_stubs_v1_stubs_service_proto_init() {
	if File_stubs_v1_stubs_service_proto != nil {
		return
	}
	file_stubs_v1_stubs_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_stubs_v1_stubs_service_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*AddStubsRequest); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*AddStubsResponse); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveStubsRequest); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveStubsResponse); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveAllStubsRequest); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveAllStubsResponse); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ListStubsRequest); i {
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
		file_stubs_v1_stubs_service_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*ListStubsResponse); i {
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
			RawDescriptor: file_stubs_v1_stubs_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_stubs_v1_stubs_service_proto_goTypes,
		DependencyIndexes: file_stubs_v1_stubs_service_proto_depIdxs,
		MessageInfos:      file_stubs_v1_stubs_service_proto_msgTypes,
	}.Build()
	File_stubs_v1_stubs_service_proto = out.File
	file_stubs_v1_stubs_service_proto_rawDesc = nil
	file_stubs_v1_stubs_service_proto_goTypes = nil
	file_stubs_v1_stubs_service_proto_depIdxs = nil
}
