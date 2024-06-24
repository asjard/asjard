// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: hello/hello.proto

package hello

import (
	_ "github.com/asjard/asjard/pkg/protobuf/annotations"
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

// 加解密示例请求
type CipherExampleReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CipherExampleReq) Reset() {
	*x = CipherExampleReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CipherExampleReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CipherExampleReq) ProtoMessage() {}

func (x *CipherExampleReq) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CipherExampleReq.ProtoReflect.Descriptor instead.
func (*CipherExampleReq) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{0}
}

type MysqlExampleReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Age  uint32 `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty"`
}

func (x *MysqlExampleReq) Reset() {
	*x = MysqlExampleReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MysqlExampleReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlExampleReq) ProtoMessage() {}

func (x *MysqlExampleReq) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlExampleReq.ProtoReflect.Descriptor instead.
func (*MysqlExampleReq) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{1}
}

func (x *MysqlExampleReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MysqlExampleReq) GetAge() uint32 {
	if x != nil {
		return x.Age
	}
	return 0
}

type MysqlExampleResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Age       uint32 `protobuf:"varint,3,opt,name=age,proto3" json:"age,omitempty"`
	CreatedAt string `protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt string `protobuf:"bytes,5,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *MysqlExampleResp) Reset() {
	*x = MysqlExampleResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MysqlExampleResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlExampleResp) ProtoMessage() {}

func (x *MysqlExampleResp) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlExampleResp.ProtoReflect.Descriptor instead.
func (*MysqlExampleResp) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{2}
}

func (x *MysqlExampleResp) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *MysqlExampleResp) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MysqlExampleResp) GetAge() uint32 {
	if x != nil {
		return x.Age
	}
	return 0
}

func (x *MysqlExampleResp) GetCreatedAt() string {
	if x != nil {
		return x.CreatedAt
	}
	return ""
}

func (x *MysqlExampleResp) GetUpdatedAt() string {
	if x != nil {
		return x.UpdatedAt
	}
	return ""
}

// 加解密示例返回
type CipherExampleResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AesEncryptValueInPlainFile         string `protobuf:"bytes,1,opt,name=aes_encrypt_value_in_plain_file,json=aesEncryptValueInPlainFile,proto3" json:"aes_encrypt_value_in_plain_file,omitempty"`
	Base64EncryptValueInPlainFile      string `protobuf:"bytes,2,opt,name=base64_encrypt_value_in_plain_file,json=base64EncryptValueInPlainFile,proto3" json:"base64_encrypt_value_in_plain_file,omitempty"`
	PlainValueInAesEncryptFile         string `protobuf:"bytes,3,opt,name=plain_value_in_aes_encrypt_file,json=plainValueInAesEncryptFile,proto3" json:"plain_value_in_aes_encrypt_file,omitempty"`
	AesEncryptValueInAesEncryptFile    string `protobuf:"bytes,4,opt,name=aes_encrypt_value_in_aes_encrypt_file,json=aesEncryptValueInAesEncryptFile,proto3" json:"aes_encrypt_value_in_aes_encrypt_file,omitempty"`
	Base64EncryptValueInAesEncryptFile string `protobuf:"bytes,5,opt,name=base64_encrypt_value_in_aes_encrypt_file,json=base64EncryptValueInAesEncryptFile,proto3" json:"base64_encrypt_value_in_aes_encrypt_file,omitempty"`
}

func (x *CipherExampleResp) Reset() {
	*x = CipherExampleResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CipherExampleResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CipherExampleResp) ProtoMessage() {}

func (x *CipherExampleResp) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CipherExampleResp.ProtoReflect.Descriptor instead.
func (*CipherExampleResp) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{3}
}

func (x *CipherExampleResp) GetAesEncryptValueInPlainFile() string {
	if x != nil {
		return x.AesEncryptValueInPlainFile
	}
	return ""
}

func (x *CipherExampleResp) GetBase64EncryptValueInPlainFile() string {
	if x != nil {
		return x.Base64EncryptValueInPlainFile
	}
	return ""
}

func (x *CipherExampleResp) GetPlainValueInAesEncryptFile() string {
	if x != nil {
		return x.PlainValueInAesEncryptFile
	}
	return ""
}

func (x *CipherExampleResp) GetAesEncryptValueInAesEncryptFile() string {
	if x != nil {
		return x.AesEncryptValueInAesEncryptFile
	}
	return ""
}

func (x *CipherExampleResp) GetBase64EncryptValueInAesEncryptFile() string {
	if x != nil {
		return x.Base64EncryptValueInAesEncryptFile
	}
	return ""
}

type SayReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RegionId  string    `protobuf:"bytes,1,opt,name=region_id,json=regionId,proto3" json:"region_id,omitempty"`
	ProjectId string    `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	UserId    int64     `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	StrList   []string  `protobuf:"bytes,4,rep,name=str_list,json=strList,proto3" json:"str_list,omitempty"`
	IntList   []int64   `protobuf:"varint,5,rep,packed,name=int_list,json=intList,proto3" json:"int_list,omitempty"`
	Obj       *SayObj   `protobuf:"bytes,6,opt,name=obj,proto3" json:"obj,omitempty"`
	Objs      []*SayObj `protobuf:"bytes,7,rep,name=objs,proto3" json:"objs,omitempty"`
}

func (x *SayReq) Reset() {
	*x = SayReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SayReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SayReq) ProtoMessage() {}

func (x *SayReq) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SayReq.ProtoReflect.Descriptor instead.
func (*SayReq) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{4}
}

func (x *SayReq) GetRegionId() string {
	if x != nil {
		return x.RegionId
	}
	return ""
}

func (x *SayReq) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *SayReq) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *SayReq) GetStrList() []string {
	if x != nil {
		return x.StrList
	}
	return nil
}

func (x *SayReq) GetIntList() []int64 {
	if x != nil {
		return x.IntList
	}
	return nil
}

func (x *SayReq) GetObj() *SayObj {
	if x != nil {
		return x.Obj
	}
	return nil
}

func (x *SayReq) GetObjs() []*SayObj {
	if x != nil {
		return x.Objs
	}
	return nil
}

type SayObj struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FieldInt int32  `protobuf:"varint,1,opt,name=field_int,json=fieldInt,proto3" json:"field_int,omitempty"`
	FieldStr string `protobuf:"bytes,2,opt,name=field_str,json=fieldStr,proto3" json:"field_str,omitempty"`
}

func (x *SayObj) Reset() {
	*x = SayObj{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SayObj) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SayObj) ProtoMessage() {}

func (x *SayObj) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SayObj.ProtoReflect.Descriptor instead.
func (*SayObj) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{5}
}

func (x *SayObj) GetFieldInt() int32 {
	if x != nil {
		return x.FieldInt
	}
	return 0
}

func (x *SayObj) GetFieldStr() string {
	if x != nil {
		return x.FieldStr
	}
	return ""
}

var File_hello_hello_proto protoreflect.FileDescriptor

var file_hello_hello_proto_rawDesc = []byte{
	0x0a, 0x11, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x1a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x73,
	0x6a, 0x61, 0x72, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x68, 0x74,
	0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0a, 0x10, 0x43, 0x69, 0x70, 0x68,
	0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x22, 0x37, 0x0a, 0x0f,
	0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x03, 0x61, 0x67, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x10, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x61, 0x67, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12,
	0x1d, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x8e,
	0x03, 0x0a, 0x11, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x12, 0x43, 0x0a, 0x1f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x70, 0x6c, 0x61,
	0x69, 0x6e, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1a, 0x61,
	0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e,
	0x50, 0x6c, 0x61, 0x69, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x49, 0x0a, 0x22, 0x62, 0x61, 0x73,
	0x65, 0x36, 0x34, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1d, 0x62, 0x61, 0x73, 0x65, 0x36, 0x34, 0x45, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x50, 0x6c, 0x61, 0x69, 0x6e,
	0x46, 0x69, 0x6c, 0x65, 0x12, 0x43, 0x0a, 0x1f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1a, 0x70,
	0x6c, 0x61, 0x69, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x41, 0x65, 0x73, 0x45, 0x6e,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x4e, 0x0a, 0x25, 0x61, 0x65, 0x73,
	0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69,
	0x6e, 0x5f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x66, 0x69,
	0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1f, 0x61, 0x65, 0x73, 0x45, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x41, 0x65, 0x73, 0x45, 0x6e,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x54, 0x0a, 0x28, 0x62, 0x61, 0x73,
	0x65, 0x36, 0x34, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x22, 0x62, 0x61, 0x73,
	0x65, 0x36, 0x34, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49,
	0x6e, 0x41, 0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x22,
	0xe5, 0x01, 0x0a, 0x06, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x65,
	0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72,
	0x65, 0x67, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12,
	0x19, 0x0a, 0x08, 0x73, 0x74, 0x72, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x74, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x6e,
	0x74, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x03, 0x52, 0x07, 0x69, 0x6e,
	0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x03, 0x6f, 0x62, 0x6a, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x2e, 0x53, 0x61, 0x79, 0x4f, 0x62, 0x6a, 0x52, 0x03, 0x6f, 0x62, 0x6a, 0x12, 0x28, 0x0a,
	0x04, 0x6f, 0x62, 0x6a, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x4f, 0x62,
	0x6a, 0x52, 0x04, 0x6f, 0x62, 0x6a, 0x73, 0x22, 0x42, 0x0a, 0x06, 0x53, 0x61, 0x79, 0x4f, 0x62,
	0x6a, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x69, 0x6e, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x49, 0x6e, 0x74, 0x12, 0x1b,
	0x0a, 0x09, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x73, 0x74, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x74, 0x72, 0x32, 0xc1, 0x03, 0x0a, 0x05,
	0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x12, 0x70, 0x0a, 0x03, 0x53, 0x61, 0x79, 0x12, 0x14, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52,
	0x65, 0x71, 0x1a, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x22, 0x3d, 0x82, 0xb5, 0x18, 0x39, 0x22, 0x37,
	0x2f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2f, 0x7b, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x64, 0x7d, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x7b, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x7b, 0x75,
	0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x12, 0x76, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12,
	0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53,
	0x61, 0x79, 0x52, 0x65, 0x71, 0x1a, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68,
	0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x22, 0x42, 0x82, 0xb5, 0x18,
	0x3e, 0x22, 0x3c, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2f,
	0x7b, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d,
	0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x7b, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x12,
	0x68, 0x0a, 0x0d, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x12, 0x1e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e,
	0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x1a, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e,
	0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x22, 0x16, 0x82, 0xb5, 0x18, 0x12, 0x12, 0x10, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x73, 0x2f, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x12, 0x64, 0x0a, 0x0c, 0x4d, 0x79, 0x73,
	0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x1d, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x1e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x22, 0x15, 0x82, 0xb5, 0x18, 0x11, 0x22, 0x0f,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x42,
	0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x73,
	0x6a, 0x61, 0x72, 0x64, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x3b, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_hello_hello_proto_rawDescOnce sync.Once
	file_hello_hello_proto_rawDescData = file_hello_hello_proto_rawDesc
)

func file_hello_hello_proto_rawDescGZIP() []byte {
	file_hello_hello_proto_rawDescOnce.Do(func() {
		file_hello_hello_proto_rawDescData = protoimpl.X.CompressGZIP(file_hello_hello_proto_rawDescData)
	})
	return file_hello_hello_proto_rawDescData
}

var file_hello_hello_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_hello_hello_proto_goTypes = []interface{}{
	(*CipherExampleReq)(nil),  // 0: api.v1.hello.CipherExampleReq
	(*MysqlExampleReq)(nil),   // 1: api.v1.hello.MysqlExampleReq
	(*MysqlExampleResp)(nil),  // 2: api.v1.hello.MysqlExampleResp
	(*CipherExampleResp)(nil), // 3: api.v1.hello.CipherExampleResp
	(*SayReq)(nil),            // 4: api.v1.hello.SayReq
	(*SayObj)(nil),            // 5: api.v1.hello.SayObj
}
var file_hello_hello_proto_depIdxs = []int32{
	5, // 0: api.v1.hello.SayReq.obj:type_name -> api.v1.hello.SayObj
	5, // 1: api.v1.hello.SayReq.objs:type_name -> api.v1.hello.SayObj
	4, // 2: api.v1.hello.Hello.Say:input_type -> api.v1.hello.SayReq
	4, // 3: api.v1.hello.Hello.Call:input_type -> api.v1.hello.SayReq
	0, // 4: api.v1.hello.Hello.CipherExample:input_type -> api.v1.hello.CipherExampleReq
	1, // 5: api.v1.hello.Hello.MysqlExample:input_type -> api.v1.hello.MysqlExampleReq
	4, // 6: api.v1.hello.Hello.Say:output_type -> api.v1.hello.SayReq
	4, // 7: api.v1.hello.Hello.Call:output_type -> api.v1.hello.SayReq
	3, // 8: api.v1.hello.Hello.CipherExample:output_type -> api.v1.hello.CipherExampleResp
	2, // 9: api.v1.hello.Hello.MysqlExample:output_type -> api.v1.hello.MysqlExampleResp
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_hello_hello_proto_init() }
func file_hello_hello_proto_init() {
	if File_hello_hello_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_hello_hello_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CipherExampleReq); i {
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
		file_hello_hello_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MysqlExampleReq); i {
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
		file_hello_hello_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MysqlExampleResp); i {
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
		file_hello_hello_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CipherExampleResp); i {
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
		file_hello_hello_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SayReq); i {
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
		file_hello_hello_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SayObj); i {
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
			RawDescriptor: file_hello_hello_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hello_hello_proto_goTypes,
		DependencyIndexes: file_hello_hello_proto_depIdxs,
		MessageInfos:      file_hello_hello_proto_msgTypes,
	}.Build()
	File_hello_hello_proto = out.File
	file_hello_hello_proto_rawDesc = nil
	file_hello_hello_proto_goTypes = nil
	file_hello_hello_proto_depIdxs = nil
}
