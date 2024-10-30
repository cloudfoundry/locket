// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: locket.proto

package models

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

type TypeCode int32

const (
	TypeCode_UNKNOWN  TypeCode = 0
	TypeCode_LOCK     TypeCode = 1
	TypeCode_PRESENCE TypeCode = 2
)

// Enum value maps for TypeCode.
var (
	TypeCode_name = map[int32]string{
		0: "UNKNOWN",
		1: "LOCK",
		2: "PRESENCE",
	}
	TypeCode_value = map[string]int32{
		"UNKNOWN":  0,
		"LOCK":     1,
		"PRESENCE": 2,
	}
)

func (x TypeCode) Enum() *TypeCode {
	p := new(TypeCode)
	*p = x
	return p
}

func (x TypeCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TypeCode) Descriptor() protoreflect.EnumDescriptor {
	return file_locket_proto_enumTypes[0].Descriptor()
}

func (TypeCode) Type() protoreflect.EnumType {
	return &file_locket_proto_enumTypes[0]
}

func (x TypeCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TypeCode.Descriptor instead.
func (TypeCode) EnumDescriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{0}
}

type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Owner string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	Value string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	// Deprecated: Marked as deprecated in locket.proto.
	Type     string   `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	TypeCode TypeCode `protobuf:"varint,5,opt,name=type_code,json=typeCode,proto3,enum=models.TypeCode" json:"type_code,omitempty"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	mi := &file_locket_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Resource) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *Resource) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// Deprecated: Marked as deprecated in locket.proto.
func (x *Resource) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Resource) GetTypeCode() TypeCode {
	if x != nil {
		return x.TypeCode
	}
	return TypeCode_UNKNOWN
}

type LockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource     *Resource `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	TtlInSeconds int64     `protobuf:"varint,2,opt,name=ttl_in_seconds,json=ttlInSeconds,proto3" json:"ttl_in_seconds,omitempty"`
}

func (x *LockRequest) Reset() {
	*x = LockRequest{}
	mi := &file_locket_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LockRequest) ProtoMessage() {}

func (x *LockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LockRequest.ProtoReflect.Descriptor instead.
func (*LockRequest) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{1}
}

func (x *LockRequest) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *LockRequest) GetTtlInSeconds() int64 {
	if x != nil {
		return x.TtlInSeconds
	}
	return 0
}

type LockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LockResponse) Reset() {
	*x = LockResponse{}
	mi := &file_locket_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LockResponse) ProtoMessage() {}

func (x *LockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LockResponse.ProtoReflect.Descriptor instead.
func (*LockResponse) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{2}
}

type ReleaseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource *Resource `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
}

func (x *ReleaseRequest) Reset() {
	*x = ReleaseRequest{}
	mi := &file_locket_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReleaseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReleaseRequest) ProtoMessage() {}

func (x *ReleaseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReleaseRequest.ProtoReflect.Descriptor instead.
func (*ReleaseRequest) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{3}
}

func (x *ReleaseRequest) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

type ReleaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReleaseResponse) Reset() {
	*x = ReleaseResponse{}
	mi := &file_locket_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReleaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReleaseResponse) ProtoMessage() {}

func (x *ReleaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReleaseResponse.ProtoReflect.Descriptor instead.
func (*ReleaseResponse) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{4}
}

type FetchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *FetchRequest) Reset() {
	*x = FetchRequest{}
	mi := &file_locket_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchRequest) ProtoMessage() {}

func (x *FetchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchRequest.ProtoReflect.Descriptor instead.
func (*FetchRequest) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{5}
}

func (x *FetchRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type FetchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource *Resource `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
}

func (x *FetchResponse) Reset() {
	*x = FetchResponse{}
	mi := &file_locket_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchResponse) ProtoMessage() {}

func (x *FetchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchResponse.ProtoReflect.Descriptor instead.
func (*FetchResponse) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{6}
}

func (x *FetchResponse) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

type FetchAllRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Deprecated: Marked as deprecated in locket.proto.
	Type     string   `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	TypeCode TypeCode `protobuf:"varint,2,opt,name=type_code,json=typeCode,proto3,enum=models.TypeCode" json:"type_code,omitempty"`
}

func (x *FetchAllRequest) Reset() {
	*x = FetchAllRequest{}
	mi := &file_locket_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchAllRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchAllRequest) ProtoMessage() {}

func (x *FetchAllRequest) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchAllRequest.ProtoReflect.Descriptor instead.
func (*FetchAllRequest) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{7}
}

// Deprecated: Marked as deprecated in locket.proto.
func (x *FetchAllRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FetchAllRequest) GetTypeCode() TypeCode {
	if x != nil {
		return x.TypeCode
	}
	return TypeCode_UNKNOWN
}

type FetchAllResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resources []*Resource `protobuf:"bytes,1,rep,name=resources,proto3" json:"resources,omitempty"`
}

func (x *FetchAllResponse) Reset() {
	*x = FetchAllResponse{}
	mi := &file_locket_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchAllResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchAllResponse) ProtoMessage() {}

func (x *FetchAllResponse) ProtoReflect() protoreflect.Message {
	mi := &file_locket_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchAllResponse.ProtoReflect.Descriptor instead.
func (*FetchAllResponse) Descriptor() ([]byte, []int) {
	return file_locket_proto_rawDescGZIP(), []int{8}
}

func (x *FetchAllResponse) GetResources() []*Resource {
	if x != nil {
		return x.Resources
	}
	return nil
}

var File_locket_proto protoreflect.FileDescriptor

var file_locket_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x22, 0x8f, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x16, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x02, 0x18, 0x01, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x2d, 0x0a, 0x09, 0x74, 0x79, 0x70,
	0x65, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x08,
	0x74, 0x79, 0x70, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x61, 0x0a, 0x0b, 0x4c, 0x6f, 0x63, 0x6b,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x74, 0x6c, 0x5f, 0x69, 0x6e, 0x5f,
	0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x74,
	0x74, 0x6c, 0x49, 0x6e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x22, 0x0e, 0x0a, 0x0c, 0x4c,
	0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3e, 0x0a, 0x0e, 0x52,
	0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x11, 0x0a, 0x0f, 0x52,
	0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x20,
	0x0a, 0x0c, 0x46, 0x65, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x22, 0x3d, 0x0a, 0x0d, 0x46, 0x65, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2c, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22,
	0x58, 0x0a, 0x0f, 0x46, 0x65, 0x74, 0x63, 0x68, 0x41, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x02, 0x18, 0x01, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x2d, 0x0a, 0x09, 0x74, 0x79,
	0x70, 0x65, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x52,
	0x08, 0x74, 0x79, 0x70, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x42, 0x0a, 0x10, 0x46, 0x65, 0x74,
	0x63, 0x68, 0x41, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a,
	0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2a, 0x2f, 0x0a,
	0x08, 0x54, 0x79, 0x70, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b,
	0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x4f, 0x43, 0x4b, 0x10, 0x01,
	0x12, 0x0c, 0x0a, 0x08, 0x50, 0x52, 0x45, 0x53, 0x45, 0x4e, 0x43, 0x45, 0x10, 0x02, 0x32, 0xf4,
	0x01, 0x0a, 0x06, 0x4c, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x33, 0x0a, 0x04, 0x4c, 0x6f, 0x63,
	0x6b, 0x12, 0x13, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x4c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x36,
	0x0a, 0x05, 0x46, 0x65, 0x74, 0x63, 0x68, 0x12, 0x14, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x07, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x12, 0x16, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x52, 0x65, 0x6c, 0x65, 0x61,
	0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x08, 0x46, 0x65, 0x74, 0x63, 0x68, 0x41, 0x6c, 0x6c,
	0x12, 0x17, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x41,
	0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x41, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x25, 0x5a, 0x23, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c,
	0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_locket_proto_rawDescOnce sync.Once
	file_locket_proto_rawDescData = file_locket_proto_rawDesc
)

func file_locket_proto_rawDescGZIP() []byte {
	file_locket_proto_rawDescOnce.Do(func() {
		file_locket_proto_rawDescData = protoimpl.X.CompressGZIP(file_locket_proto_rawDescData)
	})
	return file_locket_proto_rawDescData
}

var file_locket_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_locket_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_locket_proto_goTypes = []any{
	(TypeCode)(0),            // 0: models.TypeCode
	(*Resource)(nil),         // 1: models.Resource
	(*LockRequest)(nil),      // 2: models.LockRequest
	(*LockResponse)(nil),     // 3: models.LockResponse
	(*ReleaseRequest)(nil),   // 4: models.ReleaseRequest
	(*ReleaseResponse)(nil),  // 5: models.ReleaseResponse
	(*FetchRequest)(nil),     // 6: models.FetchRequest
	(*FetchResponse)(nil),    // 7: models.FetchResponse
	(*FetchAllRequest)(nil),  // 8: models.FetchAllRequest
	(*FetchAllResponse)(nil), // 9: models.FetchAllResponse
}
var file_locket_proto_depIdxs = []int32{
	0,  // 0: models.Resource.type_code:type_name -> models.TypeCode
	1,  // 1: models.LockRequest.resource:type_name -> models.Resource
	1,  // 2: models.ReleaseRequest.resource:type_name -> models.Resource
	1,  // 3: models.FetchResponse.resource:type_name -> models.Resource
	0,  // 4: models.FetchAllRequest.type_code:type_name -> models.TypeCode
	1,  // 5: models.FetchAllResponse.resources:type_name -> models.Resource
	2,  // 6: models.Locket.Lock:input_type -> models.LockRequest
	6,  // 7: models.Locket.Fetch:input_type -> models.FetchRequest
	4,  // 8: models.Locket.Release:input_type -> models.ReleaseRequest
	8,  // 9: models.Locket.FetchAll:input_type -> models.FetchAllRequest
	3,  // 10: models.Locket.Lock:output_type -> models.LockResponse
	7,  // 11: models.Locket.Fetch:output_type -> models.FetchResponse
	5,  // 12: models.Locket.Release:output_type -> models.ReleaseResponse
	9,  // 13: models.Locket.FetchAll:output_type -> models.FetchAllResponse
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_locket_proto_init() }
func file_locket_proto_init() {
	if File_locket_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_locket_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_locket_proto_goTypes,
		DependencyIndexes: file_locket_proto_depIdxs,
		EnumInfos:         file_locket_proto_enumTypes,
		MessageInfos:      file_locket_proto_msgTypes,
	}.Build()
	File_locket_proto = out.File
	file_locket_proto_rawDesc = nil
	file_locket_proto_goTypes = nil
	file_locket_proto_depIdxs = nil
}
