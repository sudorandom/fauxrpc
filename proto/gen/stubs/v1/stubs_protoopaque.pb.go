// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.1
// 	protoc        (unknown)
// source: stubs/v1/stubs.proto

//go:build protoopaque

package stubsv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/gofeaturespb"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ErrorCode int32

const (
	ErrorCode_ERROR_CODE_OK_UNSPECIFIED      ErrorCode = 0
	ErrorCode_ERROR_CODE_CANCELLED           ErrorCode = 1
	ErrorCode_ERROR_CODE_UNKNOWN             ErrorCode = 2
	ErrorCode_ERROR_CODE_INVALID_ARGUMENT    ErrorCode = 3
	ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED   ErrorCode = 4
	ErrorCode_ERROR_CODE_NOT_FOUND           ErrorCode = 5
	ErrorCode_ERROR_CODE_ALREADY_EXISTS      ErrorCode = 6
	ErrorCode_ERROR_CODE_PERMISSION_DENIED   ErrorCode = 7
	ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED  ErrorCode = 8
	ErrorCode_ERROR_CODE_FAILED_PRECONDITION ErrorCode = 9
	ErrorCode_ERROR_CODE_ABORTED             ErrorCode = 10
	ErrorCode_ERROR_CODE_OUT_OF_RANGE        ErrorCode = 11
	ErrorCode_ERROR_CODE_UNIMPLEMENTED       ErrorCode = 12
	ErrorCode_ERROR_CODE_INTERNAL            ErrorCode = 13
	ErrorCode_ERROR_CODE_UNAVAILABLE         ErrorCode = 14
	ErrorCode_ERROR_CODE_DATA_LOSS           ErrorCode = 15
	ErrorCode_ERROR_CODE_UNAUTHENTICATED     ErrorCode = 16
)

// Enum value maps for ErrorCode.
var (
	ErrorCode_name = map[int32]string{
		0:  "ERROR_CODE_OK_UNSPECIFIED",
		1:  "ERROR_CODE_CANCELLED",
		2:  "ERROR_CODE_UNKNOWN",
		3:  "ERROR_CODE_INVALID_ARGUMENT",
		4:  "ERROR_CODE_DEADLINE_EXCEEDED",
		5:  "ERROR_CODE_NOT_FOUND",
		6:  "ERROR_CODE_ALREADY_EXISTS",
		7:  "ERROR_CODE_PERMISSION_DENIED",
		8:  "ERROR_CODE_RESOURCE_EXHAUSTED",
		9:  "ERROR_CODE_FAILED_PRECONDITION",
		10: "ERROR_CODE_ABORTED",
		11: "ERROR_CODE_OUT_OF_RANGE",
		12: "ERROR_CODE_UNIMPLEMENTED",
		13: "ERROR_CODE_INTERNAL",
		14: "ERROR_CODE_UNAVAILABLE",
		15: "ERROR_CODE_DATA_LOSS",
		16: "ERROR_CODE_UNAUTHENTICATED",
	}
	ErrorCode_value = map[string]int32{
		"ERROR_CODE_OK_UNSPECIFIED":      0,
		"ERROR_CODE_CANCELLED":           1,
		"ERROR_CODE_UNKNOWN":             2,
		"ERROR_CODE_INVALID_ARGUMENT":    3,
		"ERROR_CODE_DEADLINE_EXCEEDED":   4,
		"ERROR_CODE_NOT_FOUND":           5,
		"ERROR_CODE_ALREADY_EXISTS":      6,
		"ERROR_CODE_PERMISSION_DENIED":   7,
		"ERROR_CODE_RESOURCE_EXHAUSTED":  8,
		"ERROR_CODE_FAILED_PRECONDITION": 9,
		"ERROR_CODE_ABORTED":             10,
		"ERROR_CODE_OUT_OF_RANGE":        11,
		"ERROR_CODE_UNIMPLEMENTED":       12,
		"ERROR_CODE_INTERNAL":            13,
		"ERROR_CODE_UNAVAILABLE":         14,
		"ERROR_CODE_DATA_LOSS":           15,
		"ERROR_CODE_UNAUTHENTICATED":     16,
	}
)

func (x ErrorCode) Enum() *ErrorCode {
	p := new(ErrorCode)
	*p = x
	return p
}

func (x ErrorCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ErrorCode) Descriptor() protoreflect.EnumDescriptor {
	return file_stubs_v1_stubs_proto_enumTypes[0].Descriptor()
}

func (ErrorCode) Type() protoreflect.EnumType {
	return &file_stubs_v1_stubs_proto_enumTypes[0]
}

func (x ErrorCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

type Stub struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Ref         *StubRef               `protobuf:"bytes,1,opt,name=ref" json:"ref,omitempty"`
	xxx_hidden_Content     isStub_Content         `protobuf_oneof:"content"`
	xxx_hidden_ActiveIf    *string                `protobuf:"bytes,5,opt,name=active_if,json=activeIf" json:"active_if,omitempty"`
	xxx_hidden_CelContent  *string                `protobuf:"bytes,6,opt,name=cel_content,json=celContent" json:"cel_content,omitempty"`
	xxx_hidden_Priority    int32                  `protobuf:"varint,7,opt,name=priority" json:"priority,omitempty"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *Stub) Reset() {
	*x = Stub{}
	mi := &file_stubs_v1_stubs_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Stub) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Stub) ProtoMessage() {}

func (x *Stub) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *Stub) GetRef() *StubRef {
	if x != nil {
		return x.xxx_hidden_Ref
	}
	return nil
}

func (x *Stub) GetProto() []byte {
	if x != nil {
		if x, ok := x.xxx_hidden_Content.(*stub_Proto); ok {
			return x.Proto
		}
	}
	return nil
}

func (x *Stub) GetJson() string {
	if x != nil {
		if x, ok := x.xxx_hidden_Content.(*stub_Json); ok {
			return x.Json
		}
	}
	return ""
}

func (x *Stub) GetError() *Error {
	if x != nil {
		if x, ok := x.xxx_hidden_Content.(*stub_Error); ok {
			return x.Error
		}
	}
	return nil
}

func (x *Stub) GetActiveIf() string {
	if x != nil {
		if x.xxx_hidden_ActiveIf != nil {
			return *x.xxx_hidden_ActiveIf
		}
		return ""
	}
	return ""
}

func (x *Stub) GetCelContent() string {
	if x != nil {
		if x.xxx_hidden_CelContent != nil {
			return *x.xxx_hidden_CelContent
		}
		return ""
	}
	return ""
}

func (x *Stub) GetPriority() int32 {
	if x != nil {
		return x.xxx_hidden_Priority
	}
	return 0
}

func (x *Stub) SetRef(v *StubRef) {
	x.xxx_hidden_Ref = v
}

func (x *Stub) SetProto(v []byte) {
	if v == nil {
		v = []byte{}
	}
	x.xxx_hidden_Content = &stub_Proto{v}
}

func (x *Stub) SetJson(v string) {
	x.xxx_hidden_Content = &stub_Json{v}
}

func (x *Stub) SetError(v *Error) {
	if v == nil {
		x.xxx_hidden_Content = nil
		return
	}
	x.xxx_hidden_Content = &stub_Error{v}
}

func (x *Stub) SetActiveIf(v string) {
	x.xxx_hidden_ActiveIf = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 2, 5)
}

func (x *Stub) SetCelContent(v string) {
	x.xxx_hidden_CelContent = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 3, 5)
}

func (x *Stub) SetPriority(v int32) {
	x.xxx_hidden_Priority = v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 4, 5)
}

func (x *Stub) HasRef() bool {
	if x == nil {
		return false
	}
	return x.xxx_hidden_Ref != nil
}

func (x *Stub) HasContent() bool {
	if x == nil {
		return false
	}
	return x.xxx_hidden_Content != nil
}

func (x *Stub) HasProto() bool {
	if x == nil {
		return false
	}
	_, ok := x.xxx_hidden_Content.(*stub_Proto)
	return ok
}

func (x *Stub) HasJson() bool {
	if x == nil {
		return false
	}
	_, ok := x.xxx_hidden_Content.(*stub_Json)
	return ok
}

func (x *Stub) HasError() bool {
	if x == nil {
		return false
	}
	_, ok := x.xxx_hidden_Content.(*stub_Error)
	return ok
}

func (x *Stub) HasActiveIf() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 2)
}

func (x *Stub) HasCelContent() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 3)
}

func (x *Stub) HasPriority() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 4)
}

func (x *Stub) ClearRef() {
	x.xxx_hidden_Ref = nil
}

func (x *Stub) ClearContent() {
	x.xxx_hidden_Content = nil
}

func (x *Stub) ClearProto() {
	if _, ok := x.xxx_hidden_Content.(*stub_Proto); ok {
		x.xxx_hidden_Content = nil
	}
}

func (x *Stub) ClearJson() {
	if _, ok := x.xxx_hidden_Content.(*stub_Json); ok {
		x.xxx_hidden_Content = nil
	}
}

func (x *Stub) ClearError() {
	if _, ok := x.xxx_hidden_Content.(*stub_Error); ok {
		x.xxx_hidden_Content = nil
	}
}

func (x *Stub) ClearActiveIf() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 2)
	x.xxx_hidden_ActiveIf = nil
}

func (x *Stub) ClearCelContent() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 3)
	x.xxx_hidden_CelContent = nil
}

func (x *Stub) ClearPriority() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 4)
	x.xxx_hidden_Priority = 0
}

const Stub_Content_not_set_case case_Stub_Content = 0
const Stub_Proto_case case_Stub_Content = 2
const Stub_Json_case case_Stub_Content = 3
const Stub_Error_case case_Stub_Content = 4

func (x *Stub) WhichContent() case_Stub_Content {
	if x == nil {
		return Stub_Content_not_set_case
	}
	switch x.xxx_hidden_Content.(type) {
	case *stub_Proto:
		return Stub_Proto_case
	case *stub_Json:
		return Stub_Json_case
	case *stub_Error:
		return Stub_Error_case
	default:
		return Stub_Content_not_set_case
	}
}

type Stub_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Ref *StubRef
	// Fields of oneof xxx_hidden_Content:
	Proto []byte
	Json  *string
	Error *Error
	// -- end of xxx_hidden_Content
	// CEL rule to decide if this stub should be used for a given request.
	ActiveIf *string
	// Similar to the json attribute but is a CEL expression that returns the result.
	CelContent *string
	Priority   *int32
}

func (b0 Stub_builder) Build() *Stub {
	m0 := &Stub{}
	b, x := &b0, m0
	_, _ = b, x
	x.xxx_hidden_Ref = b.Ref
	if b.Proto != nil {
		x.xxx_hidden_Content = &stub_Proto{b.Proto}
	}
	if b.Json != nil {
		x.xxx_hidden_Content = &stub_Json{*b.Json}
	}
	if b.Error != nil {
		x.xxx_hidden_Content = &stub_Error{b.Error}
	}
	if b.ActiveIf != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 2, 5)
		x.xxx_hidden_ActiveIf = b.ActiveIf
	}
	if b.CelContent != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 3, 5)
		x.xxx_hidden_CelContent = b.CelContent
	}
	if b.Priority != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 4, 5)
		x.xxx_hidden_Priority = *b.Priority
	}
	return m0
}

type case_Stub_Content protoreflect.FieldNumber

func (x case_Stub_Content) String() string {
	md := file_stubs_v1_stubs_proto_msgTypes[0].Descriptor()
	if x == 0 {
		return "not set"
	}
	return protoimpl.X.MessageFieldStringOf(md, protoreflect.FieldNumber(x))
}

type isStub_Content interface {
	isStub_Content()
}

type stub_Proto struct {
	Proto []byte `protobuf:"bytes,2,opt,name=proto,oneof"`
}

type stub_Json struct {
	Json string `protobuf:"bytes,3,opt,name=json,oneof"`
}

type stub_Error struct {
	Error *Error `protobuf:"bytes,4,opt,name=error,oneof"`
}

func (*stub_Proto) isStub_Content() {}

func (*stub_Json) isStub_Content() {}

func (*stub_Error) isStub_Content() {}

type StubRef struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Id          *string                `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	xxx_hidden_Target      *string                `protobuf:"bytes,2,opt,name=target" json:"target,omitempty"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *StubRef) Reset() {
	*x = StubRef{}
	mi := &file_stubs_v1_stubs_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StubRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StubRef) ProtoMessage() {}

func (x *StubRef) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *StubRef) GetId() string {
	if x != nil {
		if x.xxx_hidden_Id != nil {
			return *x.xxx_hidden_Id
		}
		return ""
	}
	return ""
}

func (x *StubRef) GetTarget() string {
	if x != nil {
		if x.xxx_hidden_Target != nil {
			return *x.xxx_hidden_Target
		}
		return ""
	}
	return ""
}

func (x *StubRef) SetId(v string) {
	x.xxx_hidden_Id = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 0, 2)
}

func (x *StubRef) SetTarget(v string) {
	x.xxx_hidden_Target = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 1, 2)
}

func (x *StubRef) HasId() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 0)
}

func (x *StubRef) HasTarget() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 1)
}

func (x *StubRef) ClearId() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 0)
	x.xxx_hidden_Id = nil
}

func (x *StubRef) ClearTarget() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 1)
	x.xxx_hidden_Target = nil
}

type StubRef_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Id     *string
	Target *string
}

func (b0 StubRef_builder) Build() *StubRef {
	m0 := &StubRef{}
	b, x := &b0, m0
	_, _ = b, x
	if b.Id != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 0, 2)
		x.xxx_hidden_Id = b.Id
	}
	if b.Target != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 1, 2)
		x.xxx_hidden_Target = b.Target
	}
	return m0
}

type Error struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Code        ErrorCode              `protobuf:"varint,1,opt,name=code,enum=stubs.v1.ErrorCode" json:"code,omitempty"`
	xxx_hidden_Message     *string                `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	xxx_hidden_Details     *[]*anypb.Any          `protobuf:"bytes,3,rep,name=details" json:"details,omitempty"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *Error) Reset() {
	*x = Error{}
	mi := &file_stubs_v1_stubs_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *Error) GetCode() ErrorCode {
	if x != nil {
		if protoimpl.X.Present(&(x.XXX_presence[0]), 0) {
			return x.xxx_hidden_Code
		}
	}
	return ErrorCode_ERROR_CODE_OK_UNSPECIFIED
}

func (x *Error) GetMessage() string {
	if x != nil {
		if x.xxx_hidden_Message != nil {
			return *x.xxx_hidden_Message
		}
		return ""
	}
	return ""
}

func (x *Error) GetDetails() []*anypb.Any {
	if x != nil {
		if x.xxx_hidden_Details != nil {
			return *x.xxx_hidden_Details
		}
	}
	return nil
}

func (x *Error) SetCode(v ErrorCode) {
	x.xxx_hidden_Code = v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 0, 3)
}

func (x *Error) SetMessage(v string) {
	x.xxx_hidden_Message = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 1, 3)
}

func (x *Error) SetDetails(v []*anypb.Any) {
	x.xxx_hidden_Details = &v
}

func (x *Error) HasCode() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 0)
}

func (x *Error) HasMessage() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 1)
}

func (x *Error) ClearCode() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 0)
	x.xxx_hidden_Code = ErrorCode_ERROR_CODE_OK_UNSPECIFIED
}

func (x *Error) ClearMessage() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 1)
	x.xxx_hidden_Message = nil
}

type Error_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Code    *ErrorCode
	Message *string
	Details []*anypb.Any
}

func (b0 Error_builder) Build() *Error {
	m0 := &Error{}
	b, x := &b0, m0
	_, _ = b, x
	if b.Code != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 0, 3)
		x.xxx_hidden_Code = *b.Code
	}
	if b.Message != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 1, 3)
		x.xxx_hidden_Message = b.Message
	}
	x.xxx_hidden_Details = &b.Details
	return m0
}

type CELGenerate struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Enabled     bool                   `protobuf:"varint,1,opt,name=enabled" json:"enabled,omitempty"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *CELGenerate) Reset() {
	*x = CELGenerate{}
	mi := &file_stubs_v1_stubs_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CELGenerate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CELGenerate) ProtoMessage() {}

func (x *CELGenerate) ProtoReflect() protoreflect.Message {
	mi := &file_stubs_v1_stubs_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *CELGenerate) GetEnabled() bool {
	if x != nil {
		return x.xxx_hidden_Enabled
	}
	return false
}

func (x *CELGenerate) SetEnabled(v bool) {
	x.xxx_hidden_Enabled = v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 0, 1)
}

func (x *CELGenerate) HasEnabled() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 0)
}

func (x *CELGenerate) ClearEnabled() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 0)
	x.xxx_hidden_Enabled = false
}

type CELGenerate_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Enabled *bool
}

func (b0 CELGenerate_builder) Build() *CELGenerate {
	m0 := &CELGenerate{}
	b, x := &b0, m0
	_, _ = b, x
	if b.Enabled != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 0, 1)
		x.xxx_hidden_Enabled = *b.Enabled
	}
	return m0
}

var File_stubs_v1_stubs_proto protoreflect.FileDescriptor

var file_stubs_v1_stubs_proto_rawDesc = []byte{
	0x0a, 0x14, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61,
	0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x6f, 0x5f, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfa, 0x01, 0x0a, 0x04,
	0x53, 0x74, 0x75, 0x62, 0x12, 0x2b, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x75,
	0x62, 0x52, 0x65, 0x66, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x03, 0x72, 0x65,
	0x66, 0x12, 0x16, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x48, 0x00, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x0a, 0x04, 0x6a, 0x73, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x12,
	0x27, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48,
	0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x63, 0x74, 0x69,
	0x76, 0x65, 0x5f, 0x69, 0x66, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x49, 0x66, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x65, 0x6c, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x65, 0x6c, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x25, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69,
	0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x42, 0x09, 0xba, 0x48, 0x06, 0x1a, 0x04, 0x18,
	0x64, 0x28, 0x00, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x42, 0x09, 0x0a,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x3e, 0x0a, 0x07, 0x53, 0x74, 0x75, 0x62,
	0x52, 0x65, 0x66, 0x12, 0x1b, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x0b, 0xba, 0x48, 0x08, 0xc8, 0x01, 0x00, 0x72, 0x03, 0x18, 0xc8, 0x01, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x89, 0x01, 0x0a, 0x05, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x12, 0x36, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x13, 0x2e, 0x73, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x43, 0x6f, 0x64, 0x65, 0x42, 0x0d, 0xba, 0x48, 0x0a, 0xc8, 0x01, 0x01, 0x82, 0x01, 0x04,
	0x10, 0x01, 0x20, 0x00, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x07, 0x64, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x22, 0x27, 0x0a, 0x0b, 0x43, 0x45, 0x4c, 0x47, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x2a, 0x83, 0x04,
	0x0a, 0x09, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1d, 0x0a, 0x19, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x4f, 0x4b, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x18, 0x0a, 0x14, 0x45, 0x52,
	0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x4c,
	0x45, 0x44, 0x10, 0x01, 0x12, 0x16, 0x0a, 0x12, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f,
	0x44, 0x45, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x02, 0x12, 0x1f, 0x0a, 0x1b,
	0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c,
	0x49, 0x44, 0x5f, 0x41, 0x52, 0x47, 0x55, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x03, 0x12, 0x20, 0x0a,
	0x1c, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x44, 0x45, 0x41, 0x44,
	0x4c, 0x49, 0x4e, 0x45, 0x5f, 0x45, 0x58, 0x43, 0x45, 0x45, 0x44, 0x45, 0x44, 0x10, 0x04, 0x12,
	0x18, 0x0a, 0x14, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x4e, 0x4f,
	0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x05, 0x12, 0x1d, 0x0a, 0x19, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x41, 0x4c, 0x52, 0x45, 0x41, 0x44, 0x59, 0x5f,
	0x45, 0x58, 0x49, 0x53, 0x54, 0x53, 0x10, 0x06, 0x12, 0x20, 0x0a, 0x1c, 0x45, 0x52, 0x52, 0x4f,
	0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x50, 0x45, 0x52, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f,
	0x4e, 0x5f, 0x44, 0x45, 0x4e, 0x49, 0x45, 0x44, 0x10, 0x07, 0x12, 0x21, 0x0a, 0x1d, 0x45, 0x52,
	0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x52, 0x45, 0x53, 0x4f, 0x55, 0x52, 0x43,
	0x45, 0x5f, 0x45, 0x58, 0x48, 0x41, 0x55, 0x53, 0x54, 0x45, 0x44, 0x10, 0x08, 0x12, 0x22, 0x0a,
	0x1e, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x46, 0x41, 0x49, 0x4c,
	0x45, 0x44, 0x5f, 0x50, 0x52, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x10,
	0x09, 0x12, 0x16, 0x0a, 0x12, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f,
	0x41, 0x42, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x10, 0x0a, 0x12, 0x1b, 0x0a, 0x17, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x4f, 0x55, 0x54, 0x5f, 0x4f, 0x46, 0x5f, 0x52,
	0x41, 0x4e, 0x47, 0x45, 0x10, 0x0b, 0x12, 0x1c, 0x0a, 0x18, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f,
	0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x4d, 0x45, 0x4e, 0x54,
	0x45, 0x44, 0x10, 0x0c, 0x12, 0x17, 0x0a, 0x13, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f,
	0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x10, 0x0d, 0x12, 0x1a, 0x0a,
	0x16, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x41, 0x56,
	0x41, 0x49, 0x4c, 0x41, 0x42, 0x4c, 0x45, 0x10, 0x0e, 0x12, 0x18, 0x0a, 0x14, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x4c, 0x4f, 0x53,
	0x53, 0x10, 0x0f, 0x12, 0x1e, 0x0a, 0x1a, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x55, 0x4e, 0x41, 0x55, 0x54, 0x48, 0x45, 0x4e, 0x54, 0x49, 0x43, 0x41, 0x54, 0x45,
	0x44, 0x10, 0x10, 0x42, 0x9d, 0x01, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x74, 0x75, 0x62,
	0x73, 0x2e, 0x76, 0x31, 0x42, 0x0a, 0x53, 0x74, 0x75, 0x62, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x75, 0x64, 0x6f, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2f, 0x66, 0x61, 0x75, 0x78, 0x72, 0x70,
	0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x74, 0x75, 0x62,
	0x73, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x74, 0x75, 0x62, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53,
	0x58, 0x58, 0xaa, 0x02, 0x08, 0x53, 0x74, 0x75, 0x62, 0x73, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x08,
	0x53, 0x74, 0x75, 0x62, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x53, 0x74, 0x75, 0x62, 0x73,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x09, 0x53, 0x74, 0x75, 0x62, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x92, 0x03, 0x05, 0xd2, 0x3e,
	0x02, 0x10, 0x02, 0x62, 0x08, 0x65, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x70, 0xe8, 0x07,
}

var file_stubs_v1_stubs_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_stubs_v1_stubs_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_stubs_v1_stubs_proto_goTypes = []any{
	(ErrorCode)(0),      // 0: stubs.v1.ErrorCode
	(*Stub)(nil),        // 1: stubs.v1.Stub
	(*StubRef)(nil),     // 2: stubs.v1.StubRef
	(*Error)(nil),       // 3: stubs.v1.Error
	(*CELGenerate)(nil), // 4: stubs.v1.CELGenerate
	(*anypb.Any)(nil),   // 5: google.protobuf.Any
}
var file_stubs_v1_stubs_proto_depIdxs = []int32{
	2, // 0: stubs.v1.Stub.ref:type_name -> stubs.v1.StubRef
	3, // 1: stubs.v1.Stub.error:type_name -> stubs.v1.Error
	0, // 2: stubs.v1.Error.code:type_name -> stubs.v1.ErrorCode
	5, // 3: stubs.v1.Error.details:type_name -> google.protobuf.Any
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_stubs_v1_stubs_proto_init() }
func file_stubs_v1_stubs_proto_init() {
	if File_stubs_v1_stubs_proto != nil {
		return
	}
	file_stubs_v1_stubs_proto_msgTypes[0].OneofWrappers = []any{
		(*stub_Proto)(nil),
		(*stub_Json)(nil),
		(*stub_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_stubs_v1_stubs_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_stubs_v1_stubs_proto_goTypes,
		DependencyIndexes: file_stubs_v1_stubs_proto_depIdxs,
		EnumInfos:         file_stubs_v1_stubs_proto_enumTypes,
		MessageInfos:      file_stubs_v1_stubs_proto_msgTypes,
	}.Build()
	File_stubs_v1_stubs_proto = out.File
	file_stubs_v1_stubs_proto_rawDesc = nil
	file_stubs_v1_stubs_proto_goTypes = nil
	file_stubs_v1_stubs_proto_depIdxs = nil
}
