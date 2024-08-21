// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: stubs/v1/stubs.proto

package stubsv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

type Stub struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ref *StubRef `protobuf:"bytes,1,opt,name=ref,proto3" json:"ref,omitempty"`
	// Types that are assignable to Content:
	//
	//	*Stub_Proto
	//	*Stub_Json
	Content isStub_Content `protobuf_oneof:"content"`
}

func (x *Stub) Reset() {
	*x = Stub{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Stub) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Stub) ProtoMessage() {}

func (x *Stub) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Stub.ProtoReflect.Descriptor instead.
func (*Stub) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_proto_rawDescGZIP(), []int{0}
}

func (x *Stub) GetRef() *StubRef {
	if x != nil {
		return x.Ref
	}
	return nil
}

func (m *Stub) GetContent() isStub_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (x *Stub) GetProto() []byte {
	if x, ok := x.GetContent().(*Stub_Proto); ok {
		return x.Proto
	}
	return nil
}

func (x *Stub) GetJson() string {
	if x, ok := x.GetContent().(*Stub_Json); ok {
		return x.Json
	}
	return ""
}

type isStub_Content interface {
	isStub_Content()
}

type Stub_Proto struct {
	Proto []byte `protobuf:"bytes,2,opt,name=proto,proto3,oneof"`
}

type Stub_Json struct {
	Json string `protobuf:"bytes,3,opt,name=json,proto3,oneof"`
}

func (*Stub_Proto) isStub_Content() {}

func (*Stub_Json) isStub_Content() {}

type StubRef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Target string `protobuf:"bytes,2,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *StubRef) Reset() {
	*x = StubRef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stubs_v1_stubs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StubRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StubRef) ProtoMessage() {}

func (x *StubRef) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StubRef.ProtoReflect.Descriptor instead.
func (*StubRef) Descriptor() ([]byte, []int) {
	return file_stubs_v1_stubs_proto_rawDescGZIP(), []int{1}
}

func (x *StubRef) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *StubRef) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

var File_stubs_v1_stubs_proto protoreflect.FileDescriptor

var file_stubs_v1_stubs_proto_rawDesc = []byte{
	0x0a, 0x14, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6c, 0x0a,
	0x04, 0x53, 0x74, 0x75, 0x62, 0x12, 0x2b, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x75, 0x62, 0x52, 0x65, 0x66, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x03, 0x72,
	0x65, 0x66, 0x12, 0x16, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x48, 0x00, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x0a, 0x04, 0x6a, 0x73,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x6a, 0x73, 0x6f, 0x6e,
	0x42, 0x09, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x40, 0x0a, 0x07, 0x53,
	0x74, 0x75, 0x62, 0x52, 0x65, 0x66, 0x12, 0x1d, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x0d, 0xba, 0x48, 0x0a, 0xc8, 0x01, 0x01, 0x72, 0x05, 0x10, 0x02, 0x18, 0xc8,
	0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x42, 0x9d, 0x01,
	0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x42, 0x0a,
	0x53, 0x74, 0x75, 0x62, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x64, 0x6f, 0x72, 0x61, 0x6e,
	0x64, 0x6f, 0x6d, 0x2f, 0x66, 0x61, 0x75, 0x78, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x69, 0x76,
	0x61, 0x74, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x74,
	0x75, 0x62, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x74, 0x75, 0x62, 0x73, 0x76, 0x31, 0xa2, 0x02,
	0x03, 0x53, 0x58, 0x58, 0xaa, 0x02, 0x08, 0x53, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x08, 0x53, 0x74, 0x75, 0x62, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x53, 0x74, 0x75,
	0x62, 0x73, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x09, 0x53, 0x74, 0x75, 0x62, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stubs_v1_stubs_proto_rawDescOnce sync.Once
	file_stubs_v1_stubs_proto_rawDescData = file_stubs_v1_stubs_proto_rawDesc
)

func file_stubs_v1_stubs_proto_rawDescGZIP() []byte {
	file_stubs_v1_stubs_proto_rawDescOnce.Do(func() {
		file_stubs_v1_stubs_proto_rawDescData = protoimpl.X.CompressGZIP(file_stubs_v1_stubs_proto_rawDescData)
	})
	return file_stubs_v1_stubs_proto_rawDescData
}

var file_stubs_v1_stubs_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_stubs_v1_stubs_proto_goTypes = []any{
	(*Stub)(nil),    // 0: stubs.v1.Stub
	(*StubRef)(nil), // 1: stubs.v1.StubRef
}
var file_stubs_v1_stubs_proto_depIdxs = []int32{
	1, // 0: stubs.v1.Stub.ref:type_name -> stubs.v1.StubRef
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_stubs_v1_stubs_proto_init() }
func file_stubs_v1_stubs_proto_init() {
	if File_stubs_v1_stubs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_stubs_v1_stubs_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Stub); i {
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
		file_stubs_v1_stubs_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*StubRef); i {
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
	file_stubs_v1_stubs_proto_msgTypes[0].OneofWrappers = []any{
		(*Stub_Proto)(nil),
		(*Stub_Json)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_stubs_v1_stubs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_stubs_v1_stubs_proto_goTypes,
		DependencyIndexes: file_stubs_v1_stubs_proto_depIdxs,
		MessageInfos:      file_stubs_v1_stubs_proto_msgTypes,
	}.Build()
	File_stubs_v1_stubs_proto = out.File
	file_stubs_v1_stubs_proto_rawDesc = nil
	file_stubs_v1_stubs_proto_goTypes = nil
	file_stubs_v1_stubs_proto_depIdxs = nil
}
