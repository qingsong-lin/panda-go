// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.1
// source: common.proto

package pb

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type DownloadPprofFileRequest_PprofType int32

const (
	DownloadPprofFileRequest_goroutine    DownloadPprofFileRequest_PprofType = 0
	DownloadPprofFileRequest_heap         DownloadPprofFileRequest_PprofType = 1
	DownloadPprofFileRequest_threadcreate DownloadPprofFileRequest_PprofType = 2
	DownloadPprofFileRequest_allocs       DownloadPprofFileRequest_PprofType = 3
	DownloadPprofFileRequest_block        DownloadPprofFileRequest_PprofType = 4
	DownloadPprofFileRequest_mutex        DownloadPprofFileRequest_PprofType = 5
	DownloadPprofFileRequest_profile      DownloadPprofFileRequest_PprofType = 6
)

// Enum value maps for DownloadPprofFileRequest_PprofType.
var (
	DownloadPprofFileRequest_PprofType_name = map[int32]string{
		0: "goroutine",
		1: "heap",
		2: "threadcreate",
		3: "allocs",
		4: "block",
		5: "mutex",
		6: "profile",
	}
	DownloadPprofFileRequest_PprofType_value = map[string]int32{
		"goroutine":    0,
		"heap":         1,
		"threadcreate": 2,
		"allocs":       3,
		"block":        4,
		"mutex":        5,
		"profile":      6,
	}
)

func (x DownloadPprofFileRequest_PprofType) Enum() *DownloadPprofFileRequest_PprofType {
	p := new(DownloadPprofFileRequest_PprofType)
	*p = x
	return p
}

func (x DownloadPprofFileRequest_PprofType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DownloadPprofFileRequest_PprofType) Descriptor() protoreflect.EnumDescriptor {
	return file_common_proto_enumTypes[0].Descriptor()
}

func (DownloadPprofFileRequest_PprofType) Type() protoreflect.EnumType {
	return &file_common_proto_enumTypes[0]
}

func (x DownloadPprofFileRequest_PprofType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DownloadPprofFileRequest_PprofType.Descriptor instead.
func (DownloadPprofFileRequest_PprofType) EnumDescriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0, 0}
}

type DownloadPprofFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PprofTypes []DownloadPprofFileRequest_PprofType `protobuf:"varint,1,rep,packed,name=pprof_types,json=pprofTypes,proto3,enum=pb.DownloadPprofFileRequest_PprofType" json:"pprof_types,omitempty"`
}

func (x *DownloadPprofFileRequest) Reset() {
	*x = DownloadPprofFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadPprofFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadPprofFileRequest) ProtoMessage() {}

func (x *DownloadPprofFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadPprofFileRequest.ProtoReflect.Descriptor instead.
func (*DownloadPprofFileRequest) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0}
}

func (x *DownloadPprofFileRequest) GetPprofTypes() []DownloadPprofFileRequest_PprofType {
	if x != nil {
		return x.PprofTypes
	}
	return nil
}

type FileChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *FileChunk) Reset() {
	*x = FileChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChunk) ProtoMessage() {}

func (x *FileChunk) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChunk.ProtoReflect.Descriptor instead.
func (*FileChunk) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{1}
}

func (x *FileChunk) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type SetLogLevelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Level string `protobuf:"bytes,1,opt,name=level,proto3" json:"level,omitempty"`
}

func (x *SetLogLevelRequest) Reset() {
	*x = SetLogLevelRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetLogLevelRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetLogLevelRequest) ProtoMessage() {}

func (x *SetLogLevelRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetLogLevelRequest.ProtoReflect.Descriptor instead.
func (*SetLogLevelRequest) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{2}
}

func (x *SetLogLevelRequest) GetLevel() string {
	if x != nil {
		return x.Level
	}
	return ""
}

type SetLogLevelResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code   int32  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	ErrMsg string `protobuf:"bytes,2,opt,name=err_msg,json=errMsg,proto3" json:"err_msg,omitempty"`
}

func (x *SetLogLevelResponse) Reset() {
	*x = SetLogLevelResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetLogLevelResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetLogLevelResponse) ProtoMessage() {}

func (x *SetLogLevelResponse) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetLogLevelResponse.ProtoReflect.Descriptor instead.
func (*SetLogLevelResponse) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{3}
}

func (x *SetLogLevelResponse) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *SetLogLevelResponse) GetErrMsg() string {
	if x != nil {
		return x.ErrMsg
	}
	return ""
}

type HelloRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *HelloRequest) Reset() {
	*x = HelloRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HelloRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloRequest) ProtoMessage() {}

func (x *HelloRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloRequest.ProtoReflect.Descriptor instead.
func (*HelloRequest) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{4}
}

func (x *HelloRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// 响应消息
type HelloResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Reply string `protobuf:"bytes,1,opt,name=reply,proto3" json:"reply,omitempty"`
}

func (x *HelloResponse) Reset() {
	*x = HelloResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HelloResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloResponse) ProtoMessage() {}

func (x *HelloResponse) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloResponse.ProtoReflect.Descriptor instead.
func (*HelloResponse) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{5}
}

func (x *HelloResponse) GetReply() string {
	if x != nil {
		return x.Reply
	}
	return ""
}

var File_common_proto protoreflect.FileDescriptor

var file_common_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02,
	0x70, 0x62, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f,
	0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xca, 0x01, 0x0a, 0x18, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x70, 0x72,
	0x6f, 0x66, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x47, 0x0a,
	0x0b, 0x70, 0x70, 0x72, 0x6f, 0x66, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0e, 0x32, 0x26, 0x2e, 0x70, 0x62, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64,
	0x50, 0x70, 0x72, 0x6f, 0x66, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x2e, 0x50, 0x70, 0x72, 0x6f, 0x66, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x70, 0x70, 0x72, 0x6f,
	0x66, 0x54, 0x79, 0x70, 0x65, 0x73, 0x22, 0x65, 0x0a, 0x09, 0x50, 0x70, 0x72, 0x6f, 0x66, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x67, 0x6f, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x65,
	0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x70, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c,
	0x74, 0x68, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x10, 0x02, 0x12, 0x0a,
	0x0a, 0x06, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x73, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x10, 0x04, 0x12, 0x09, 0x0a, 0x05, 0x6d, 0x75, 0x74, 0x65, 0x78, 0x10, 0x05,
	0x12, 0x0b, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x10, 0x06, 0x22, 0x1f, 0x0a,
	0x09, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x2a,
	0x0a, 0x12, 0x73, 0x65, 0x74, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x22, 0x42, 0x0a, 0x13, 0x53, 0x65,
	0x74, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x65, 0x72, 0x72, 0x5f, 0x6d, 0x73, 0x67,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x72, 0x72, 0x4d, 0x73, 0x67, 0x22, 0x22,
	0x0a, 0x0c, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x22, 0x25, 0x0a, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x32, 0xc7, 0x01, 0x0a, 0x0d, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x53, 0x0a, 0x0b, 0x73,
	0x65, 0x74, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x16, 0x2e, 0x70, 0x62, 0x2e,
	0x73, 0x65, 0x74, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x74, 0x4c, 0x6f, 0x67, 0x4c, 0x65,
	0x76, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x0d, 0x12, 0x0b, 0x2f, 0x76, 0x31, 0x2f, 0x7b, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x7d,
	0x12, 0x61, 0x0a, 0x11, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x70, 0x72, 0x6f,
	0x66, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1c, 0x2e, 0x70, 0x62, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c,
	0x6f, 0x61, 0x64, 0x50, 0x70, 0x72, 0x6f, 0x66, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75,
	0x6e, 0x6b, 0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x22, 0x12, 0x2f, 0x76, 0x31, 0x2f,
	0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x2d, 0x70, 0x70, 0x72, 0x6f, 0x66, 0x3a, 0x01,
	0x2a, 0x30, 0x01, 0x42, 0x58, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x92, 0x41, 0x4f, 0x12, 0x26,
	0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x20, 0x68, 0x74, 0x74, 0x70, 0x20, 0x61, 0x70,
	0x69, 0x2a, 0x0e, 0x0a, 0x0c, 0x4e, 0x4f, 0x2e, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x01, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_proto_rawDescOnce sync.Once
	file_common_proto_rawDescData = file_common_proto_rawDesc
)

func file_common_proto_rawDescGZIP() []byte {
	file_common_proto_rawDescOnce.Do(func() {
		file_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_proto_rawDescData)
	})
	return file_common_proto_rawDescData
}

var file_common_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_common_proto_goTypes = []interface{}{
	(DownloadPprofFileRequest_PprofType)(0), // 0: pb.DownloadPprofFileRequest.PprofType
	(*DownloadPprofFileRequest)(nil),        // 1: pb.DownloadPprofFileRequest
	(*FileChunk)(nil),                       // 2: pb.FileChunk
	(*SetLogLevelRequest)(nil),              // 3: pb.setLogLevelRequest
	(*SetLogLevelResponse)(nil),             // 4: pb.SetLogLevelResponse
	(*HelloRequest)(nil),                    // 5: pb.HelloRequest
	(*HelloResponse)(nil),                   // 6: pb.HelloResponse
}
var file_common_proto_depIdxs = []int32{
	0, // 0: pb.DownloadPprofFileRequest.pprof_types:type_name -> pb.DownloadPprofFileRequest.PprofType
	3, // 1: pb.CommonService.setLogLevel:input_type -> pb.setLogLevelRequest
	1, // 2: pb.CommonService.DownloadPprofFile:input_type -> pb.DownloadPprofFileRequest
	4, // 3: pb.CommonService.setLogLevel:output_type -> pb.SetLogLevelResponse
	2, // 4: pb.CommonService.DownloadPprofFile:output_type -> pb.FileChunk
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_common_proto_init() }
func file_common_proto_init() {
	if File_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadPprofFileRequest); i {
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
		file_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileChunk); i {
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
		file_common_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetLogLevelRequest); i {
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
		file_common_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetLogLevelResponse); i {
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
		file_common_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HelloRequest); i {
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
		file_common_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HelloResponse); i {
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
			RawDescriptor: file_common_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_common_proto_goTypes,
		DependencyIndexes: file_common_proto_depIdxs,
		EnumInfos:         file_common_proto_enumTypes,
		MessageInfos:      file_common_proto_msgTypes,
	}.Build()
	File_common_proto = out.File
	file_common_proto_rawDesc = nil
	file_common_proto_goTypes = nil
	file_common_proto_depIdxs = nil
}
