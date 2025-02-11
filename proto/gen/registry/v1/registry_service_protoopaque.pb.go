// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.1
// 	protoc        (unknown)
// source: registry/v1/registry_service.proto

//go:build protoopaque

package registryv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	_ "google.golang.org/protobuf/types/gofeaturespb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AddDescriptorsRequest struct {
	state                  protoimpl.MessageState          `protogen:"opaque.v1"`
	xxx_hidden_Descriptors *descriptorpb.FileDescriptorSet `protobuf:"bytes,1,opt,name=descriptors" json:"descriptors,omitempty"`
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *AddDescriptorsRequest) Reset() {
	*x = AddDescriptorsRequest{}
	mi := &file_registry_v1_registry_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddDescriptorsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddDescriptorsRequest) ProtoMessage() {}

func (x *AddDescriptorsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_registry_v1_registry_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *AddDescriptorsRequest) GetDescriptors() *descriptorpb.FileDescriptorSet {
	if x != nil {
		return x.xxx_hidden_Descriptors
	}
	return nil
}

func (x *AddDescriptorsRequest) SetDescriptors(v *descriptorpb.FileDescriptorSet) {
	x.xxx_hidden_Descriptors = v
}

func (x *AddDescriptorsRequest) HasDescriptors() bool {
	if x == nil {
		return false
	}
	return x.xxx_hidden_Descriptors != nil
}

func (x *AddDescriptorsRequest) ClearDescriptors() {
	x.xxx_hidden_Descriptors = nil
}

type AddDescriptorsRequest_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Descriptors *descriptorpb.FileDescriptorSet
}

func (b0 AddDescriptorsRequest_builder) Build() *AddDescriptorsRequest {
	m0 := &AddDescriptorsRequest{}
	b, x := &b0, m0
	_, _ = b, x
	x.xxx_hidden_Descriptors = b.Descriptors
	return m0
}

type AddDescriptorsResponse struct {
	state         protoimpl.MessageState `protogen:"opaque.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddDescriptorsResponse) Reset() {
	*x = AddDescriptorsResponse{}
	mi := &file_registry_v1_registry_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddDescriptorsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddDescriptorsResponse) ProtoMessage() {}

func (x *AddDescriptorsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_registry_v1_registry_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

type AddDescriptorsResponse_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

}

func (b0 AddDescriptorsResponse_builder) Build() *AddDescriptorsResponse {
	m0 := &AddDescriptorsResponse{}
	b, x := &b0, m0
	_, _ = b, x
	return m0
}

type ResetRequest struct {
	state         protoimpl.MessageState `protogen:"opaque.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ResetRequest) Reset() {
	*x = ResetRequest{}
	mi := &file_registry_v1_registry_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResetRequest) ProtoMessage() {}

func (x *ResetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_registry_v1_registry_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

type ResetRequest_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

}

func (b0 ResetRequest_builder) Build() *ResetRequest {
	m0 := &ResetRequest{}
	b, x := &b0, m0
	_, _ = b, x
	return m0
}

type ResetResponse struct {
	state         protoimpl.MessageState `protogen:"opaque.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ResetResponse) Reset() {
	*x = ResetResponse{}
	mi := &file_registry_v1_registry_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResetResponse) ProtoMessage() {}

func (x *ResetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_registry_v1_registry_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

type ResetResponse_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

}

func (b0 ResetResponse_builder) Build() *ResetResponse {
	m0 := &ResetResponse{}
	b, x := &b0, m0
	_, _ = b, x
	return m0
}

var File_registry_v1_registry_service_proto protoreflect.FileDescriptor

var file_registry_v1_registry_service_proto_rawDesc = []byte{
	0x0a, 0x22, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x6f, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5d, 0x0a, 0x15, 0x41, 0x64, 0x64, 0x44, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x44, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x74, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x73, 0x22, 0x18, 0x0a, 0x16, 0x41, 0x64, 0x64, 0x44, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x0e, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x0f, 0x0a, 0x0d, 0x52, 0x65, 0x73, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x32, 0xb0, 0x01, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x5b, 0x0a, 0x0e, 0x41, 0x64, 0x64, 0x44, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x22, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x44, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x40, 0x0a, 0x05, 0x52, 0x65, 0x73, 0x65, 0x74, 0x12, 0x19, 0x2e, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x65, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0xbc, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x14, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x64, 0x6f,
	0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2f, 0x66, 0x61, 0x75, 0x78, 0x72, 0x70, 0x63, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x76, 0x31, 0xa2,
	0x02, 0x03, 0x52, 0x58, 0x58, 0xaa, 0x02, 0x0b, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x56, 0x31, 0xca, 0x02, 0x0b, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5c, 0x56,
	0x31, 0xe2, 0x02, 0x17, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0c, 0x52, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x3a, 0x3a, 0x56, 0x31, 0x92, 0x03, 0x05, 0xd2, 0x3e, 0x02,
	0x10, 0x02, 0x62, 0x08, 0x65, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x70, 0xe8, 0x07,
}

var file_registry_v1_registry_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_registry_v1_registry_service_proto_goTypes = []any{
	(*AddDescriptorsRequest)(nil),          // 0: registry.v1.AddDescriptorsRequest
	(*AddDescriptorsResponse)(nil),         // 1: registry.v1.AddDescriptorsResponse
	(*ResetRequest)(nil),                   // 2: registry.v1.ResetRequest
	(*ResetResponse)(nil),                  // 3: registry.v1.ResetResponse
	(*descriptorpb.FileDescriptorSet)(nil), // 4: google.protobuf.FileDescriptorSet
}
var file_registry_v1_registry_service_proto_depIdxs = []int32{
	4, // 0: registry.v1.AddDescriptorsRequest.descriptors:type_name -> google.protobuf.FileDescriptorSet
	0, // 1: registry.v1.RegistryService.AddDescriptors:input_type -> registry.v1.AddDescriptorsRequest
	2, // 2: registry.v1.RegistryService.Reset:input_type -> registry.v1.ResetRequest
	1, // 3: registry.v1.RegistryService.AddDescriptors:output_type -> registry.v1.AddDescriptorsResponse
	3, // 4: registry.v1.RegistryService.Reset:output_type -> registry.v1.ResetResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_registry_v1_registry_service_proto_init() }
func file_registry_v1_registry_service_proto_init() {
	if File_registry_v1_registry_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_registry_v1_registry_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_registry_v1_registry_service_proto_goTypes,
		DependencyIndexes: file_registry_v1_registry_service_proto_depIdxs,
		MessageInfos:      file_registry_v1_registry_service_proto_msgTypes,
	}.Build()
	File_registry_v1_registry_service_proto = out.File
	file_registry_v1_registry_service_proto_rawDesc = nil
	file_registry_v1_registry_service_proto_goTypes = nil
	file_registry_v1_registry_service_proto_depIdxs = nil
}
