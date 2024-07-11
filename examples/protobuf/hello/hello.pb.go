// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: hello/hello.proto

package hello

import (
	_ "github.com/asjard/asjard/pkg/protobuf/httppb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Kind int32

const (
	Kind_K_A Kind = 0
	Kind_K_B Kind = 1
)

// Enum value maps for Kind.
var (
	Kind_name = map[int32]string{
		0: "K_A",
		1: "K_B",
	}
	Kind_value = map[string]int32{
		"K_A": 0,
		"K_B": 1,
	}
)

func (x Kind) Enum() *Kind {
	p := new(Kind)
	*p = x
	return p
}

func (x Kind) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Kind) Descriptor() protoreflect.EnumDescriptor {
	return file_hello_hello_proto_enumTypes[0].Descriptor()
}

func (Kind) Type() protoreflect.EnumType {
	return &file_hello_hello_proto_enumTypes[0]
}

func (x Kind) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Kind.Descriptor instead.
func (Kind) EnumDescriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{0}
}

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

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	Age  uint32 `protobuf:"varint,2,opt,name=age,proto3" json:"age"`
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

	Id        int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`
	Age       uint32 `protobuf:"varint,3,opt,name=age,proto3" json:"age"`
	CreatedAt string `protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3" json:"created_at"`
	UpdatedAt string `protobuf:"bytes,5,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at"`
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

	AesEncryptValueInPlainFile         string `protobuf:"bytes,1,opt,name=aes_encrypt_value_in_plain_file,json=aesEncryptValueInPlainFile,proto3" json:"aes_encrypt_value_in_plain_file"`
	Base64EncryptValueInPlainFile      string `protobuf:"bytes,2,opt,name=base64_encrypt_value_in_plain_file,json=base64EncryptValueInPlainFile,proto3" json:"base64_encrypt_value_in_plain_file"`
	PlainValueInAesEncryptFile         string `protobuf:"bytes,3,opt,name=plain_value_in_aes_encrypt_file,json=plainValueInAesEncryptFile,proto3" json:"plain_value_in_aes_encrypt_file"`
	AesEncryptValueInAesEncryptFile    string `protobuf:"bytes,4,opt,name=aes_encrypt_value_in_aes_encrypt_file,json=aesEncryptValueInAesEncryptFile,proto3" json:"aes_encrypt_value_in_aes_encrypt_file"`
	Base64EncryptValueInAesEncryptFile string `protobuf:"bytes,5,opt,name=base64_encrypt_value_in_aes_encrypt_file,json=base64EncryptValueInAesEncryptFile,proto3" json:"base64_encrypt_value_in_aes_encrypt_file"`
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

	// 区域ID
	RegionId string `protobuf:"bytes,1,opt,name=region_id,json=regionId,proto3" json:"region_id"`
	// 项目ID
	ProjectId string `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id"`
	// 用户ID
	UserId int64 `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id"`
	// 字符串列表
	StrList []string `protobuf:"bytes,4,rep,name=str_list,json=strList,proto3" json:"str_list"`
	// 数字列表
	IntList []int64 `protobuf:"varint,5,rep,packed,name=int_list,json=intList,proto3" json:"int_list"`
	// 对象
	Obj *SayObj `protobuf:"bytes,6,opt,name=obj,proto3" json:"obj"`
	// 对象列表
	Objs []*SayObj `protobuf:"bytes,7,rep,name=objs,proto3" json:"objs"`
	// 配置
	Configs *Configs `protobuf:"bytes,8,opt,name=configs,proto3" json:"configs"`
	// 分页
	Page int32 `protobuf:"varint,9,opt,name=page,proto3" json:"page"`
	// 每页大小
	Size int32 `protobuf:"varint,10,opt,name=size,proto3" json:"size"`
	// 排序
	Sort string `protobuf:"bytes,11,opt,name=sort,proto3" json:"sort"`
	// 布尔类型
	Ok *bool `protobuf:"varint,12,opt,name=ok,proto3,oneof" json:"ok"`
	// 可选整形参数
	IntOptionalValue *int32 `protobuf:"varint,13,opt,name=int_optional_value,json=intOptionalValue,proto3,oneof" json:"int_optional_value"`
	// 可选字符串参数
	StringOptionalValue *string `protobuf:"bytes,14,opt,name=string_optional_value,json=stringOptionalValue,proto3,oneof" json:"string_optional_value"`
	// 可选枚举参数
	Kind *Kind `protobuf:"varint,15,opt,name=kind,proto3,enum=api.v1.hello.Kind,oneof" json:"kind"`
	// 枚举列表
	Kinds      []Kind `protobuf:"varint,16,rep,packed,name=kinds,proto3,enum=api.v1.hello.Kind" json:"kinds"`
	BytesValue []byte `protobuf:"bytes,17,opt,name=bytes_value,json=bytesValue,proto3" json:"bytes_value"`
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

func (x *SayReq) GetConfigs() *Configs {
	if x != nil {
		return x.Configs
	}
	return nil
}

func (x *SayReq) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *SayReq) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *SayReq) GetSort() string {
	if x != nil {
		return x.Sort
	}
	return ""
}

func (x *SayReq) GetOk() bool {
	if x != nil && x.Ok != nil {
		return *x.Ok
	}
	return false
}

func (x *SayReq) GetIntOptionalValue() int32 {
	if x != nil && x.IntOptionalValue != nil {
		return *x.IntOptionalValue
	}
	return 0
}

func (x *SayReq) GetStringOptionalValue() string {
	if x != nil && x.StringOptionalValue != nil {
		return *x.StringOptionalValue
	}
	return ""
}

func (x *SayReq) GetKind() Kind {
	if x != nil && x.Kind != nil {
		return *x.Kind
	}
	return Kind_K_A
}

func (x *SayReq) GetKinds() []Kind {
	if x != nil {
		return x.Kinds
	}
	return nil
}

func (x *SayReq) GetBytesValue() []byte {
	if x != nil {
		return x.BytesValue
	}
	return nil
}

type Configs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timeout                                     string `protobuf:"bytes,1,opt,name=timeout,proto3" json:"timeout"`
	FieldInDifferentFileUnderSameSection        string `protobuf:"bytes,2,opt,name=field_in_different_file_under_same_section,json=fieldInDifferentFileUnderSameSection,proto3" json:"field_in_different_file_under_same_section"`
	AnotherFieldInDifferentFileUnderSameSection string `protobuf:"bytes,3,opt,name=another_field_in_different_file_under_same_section,json=anotherFieldInDifferentFileUnderSameSection,proto3" json:"another_field_in_different_file_under_same_section"`
}

func (x *Configs) Reset() {
	*x = Configs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Configs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Configs) ProtoMessage() {}

func (x *Configs) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Configs.ProtoReflect.Descriptor instead.
func (*Configs) Descriptor() ([]byte, []int) {
	return file_hello_hello_proto_rawDescGZIP(), []int{5}
}

func (x *Configs) GetTimeout() string {
	if x != nil {
		return x.Timeout
	}
	return ""
}

func (x *Configs) GetFieldInDifferentFileUnderSameSection() string {
	if x != nil {
		return x.FieldInDifferentFileUnderSameSection
	}
	return ""
}

func (x *Configs) GetAnotherFieldInDifferentFileUnderSameSection() string {
	if x != nil {
		return x.AnotherFieldInDifferentFileUnderSameSection
	}
	return ""
}

type SayObj struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FieldInt int32  `protobuf:"varint,1,opt,name=field_int,json=fieldInt,proto3" json:"field_int"`
	FieldStr string `protobuf:"bytes,2,opt,name=field_str,json=fieldStr,proto3" json:"field_str"`
}

func (x *SayObj) Reset() {
	*x = SayObj{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hello_hello_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SayObj) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SayObj) ProtoMessage() {}

func (x *SayObj) ProtoReflect() protoreflect.Message {
	mi := &file_hello_hello_proto_msgTypes[6]
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
	return file_hello_hello_proto_rawDescGZIP(), []int{6}
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
	0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0a, 0x10, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x22, 0x37, 0x0a, 0x0f, 0x4d, 0x79, 0x73,
	0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x61,
	0x67, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x10, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x61,
	0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x61, 0x67, 0x65, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x8e, 0x03, 0x0a, 0x11,
	0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x43, 0x0a, 0x1f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x5f,
	0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1a, 0x61, 0x65, 0x73, 0x45,
	0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x50, 0x6c, 0x61,
	0x69, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x49, 0x0a, 0x22, 0x62, 0x61, 0x73, 0x65, 0x36, 0x34,
	0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69,
	0x6e, 0x5f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x1d, 0x62, 0x61, 0x73, 0x65, 0x36, 0x34, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x50, 0x6c, 0x61, 0x69, 0x6e, 0x46, 0x69, 0x6c,
	0x65, 0x12, 0x43, 0x0a, 0x1f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x5f, 0x69, 0x6e, 0x5f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f,
	0x66, 0x69, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1a, 0x70, 0x6c, 0x61, 0x69,
	0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x41, 0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x4e, 0x0a, 0x25, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x61,
	0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1f, 0x61, 0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x41, 0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x54, 0x0a, 0x28, 0x62, 0x61, 0x73, 0x65, 0x36, 0x34,
	0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69,
	0x6e, 0x5f, 0x61, 0x65, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x66, 0x69,
	0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x22, 0x62, 0x61, 0x73, 0x65, 0x36, 0x34,
	0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x41, 0x65,
	0x73, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x22, 0x8c, 0x05, 0x0a,
	0x06, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x65, 0x67, 0x69, 0x6f,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x67, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08,
	0x73, 0x74, 0x72, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x73, 0x74, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x5f, 0x6c,
	0x69, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x03, 0x52, 0x07, 0x69, 0x6e, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x26, 0x0a, 0x03, 0x6f, 0x62, 0x6a, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53,
	0x61, 0x79, 0x4f, 0x62, 0x6a, 0x52, 0x03, 0x6f, 0x62, 0x6a, 0x12, 0x28, 0x0a, 0x04, 0x6f, 0x62,
	0x6a, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x4f, 0x62, 0x6a, 0x52, 0x04,
	0x6f, 0x62, 0x6a, 0x73, 0x12, 0x2f, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68,
	0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73, 0x6f, 0x72,
	0x74, 0x12, 0x13, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52,
	0x02, 0x6f, 0x6b, 0x88, 0x01, 0x01, 0x12, 0x31, 0x0a, 0x12, 0x69, 0x6e, 0x74, 0x5f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x0d, 0x20, 0x01,
	0x28, 0x05, 0x48, 0x01, 0x52, 0x10, 0x69, 0x6e, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61,
	0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x88, 0x01, 0x01, 0x12, 0x37, 0x0a, 0x15, 0x73, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x13, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x2b, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e,
	0x4b, 0x69, 0x6e, 0x64, 0x48, 0x03, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x88, 0x01, 0x01, 0x12,
	0x28, 0x0a, 0x05, 0x6b, 0x69, 0x6e, 0x64, 0x73, 0x18, 0x10, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x12,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x4b, 0x69,
	0x6e, 0x64, 0x52, 0x05, 0x6b, 0x69, 0x6e, 0x64, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x79, 0x74,
	0x65, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a,
	0x62, 0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x6f,
	0x6b, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x69, 0x6e, 0x74, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x61, 0x6c, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x18, 0x0a, 0x16, 0x5f, 0x73, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x6b, 0x69, 0x6e, 0x64, 0x22, 0xe6, 0x01, 0x0a, 0x07,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75,
	0x74, 0x12, 0x58, 0x0a, 0x2a, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x69, 0x6e, 0x5f, 0x64, 0x69,
	0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x75, 0x6e, 0x64,
	0x65, 0x72, 0x5f, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x24, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x49, 0x6e, 0x44, 0x69,
	0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x55, 0x6e, 0x64, 0x65, 0x72,
	0x53, 0x61, 0x6d, 0x65, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x67, 0x0a, 0x32, 0x61,
	0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x69, 0x6e, 0x5f,
	0x64, 0x69, 0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x75,
	0x6e, 0x64, 0x65, 0x72, 0x5f, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x2b, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x49, 0x6e, 0x44, 0x69, 0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74,
	0x46, 0x69, 0x6c, 0x65, 0x55, 0x6e, 0x64, 0x65, 0x72, 0x53, 0x61, 0x6d, 0x65, 0x53, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x42, 0x0a, 0x06, 0x53, 0x61, 0x79, 0x4f, 0x62, 0x6a, 0x12, 0x1b,
	0x0a, 0x09, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x49, 0x6e, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x66,
	0x69, 0x65, 0x6c, 0x64, 0x5f, 0x73, 0x74, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x74, 0x72, 0x2a, 0x18, 0x0a, 0x04, 0x4b, 0x69, 0x6e, 0x64,
	0x12, 0x07, 0x0a, 0x03, 0x4b, 0x5f, 0x41, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x4b, 0x5f, 0x42,
	0x10, 0x01, 0x32, 0xc2, 0x04, 0x0a, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x12, 0xad, 0x01, 0x0a,
	0x03, 0x53, 0x61, 0x79, 0x12, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x1a, 0x14, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71,
	0x22, 0x7a, 0x82, 0xb5, 0x18, 0x39, 0x22, 0x37, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2f,
	0x7b, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d,
	0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x7b, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x82,
	0xb5, 0x18, 0x39, 0x12, 0x37, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2f, 0x7b, 0x72, 0x65,
	0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x75, 0x73,
	0x65, 0x72, 0x2f, 0x7b, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x12, 0x76, 0x0a, 0x04,
	0x43, 0x61, 0x6c, 0x6c, 0x12, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71, 0x1a, 0x14, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x53, 0x61, 0x79, 0x52, 0x65, 0x71,
	0x22, 0x42, 0x82, 0xb5, 0x18, 0x3e, 0x22, 0x3c, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x6f, 0x6e, 0x2f, 0x7b, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x7d,
	0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x7b, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x7d, 0x12, 0x41, 0x0a, 0x03, 0x4c, 0x6f, 0x67, 0x12, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x0a, 0x82, 0xb5, 0x18,
	0x06, 0x12, 0x04, 0x2f, 0x6c, 0x6f, 0x67, 0x12, 0x68, 0x0a, 0x0d, 0x43, 0x69, 0x70, 0x68, 0x65,
	0x72, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x1e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x22, 0x16, 0x82, 0xb5, 0x18, 0x12, 0x12,
	0x10, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x63, 0x69, 0x70, 0x68, 0x65,
	0x72, 0x12, 0x64, 0x0a, 0x0c, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x12, 0x1d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
	0x2e, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x1a, 0x1e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e,
	0x4d, 0x79, 0x73, 0x71, 0x6c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x22, 0x15, 0x82, 0xb5, 0x18, 0x11, 0x22, 0x0f, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x73, 0x2f, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x61, 0x73, 0x6a,
	0x61, 0x72, 0x64, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x3b, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_hello_hello_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_hello_hello_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_hello_hello_proto_goTypes = []interface{}{
	(Kind)(0),                 // 0: api.v1.hello.Kind
	(*CipherExampleReq)(nil),  // 1: api.v1.hello.CipherExampleReq
	(*MysqlExampleReq)(nil),   // 2: api.v1.hello.MysqlExampleReq
	(*MysqlExampleResp)(nil),  // 3: api.v1.hello.MysqlExampleResp
	(*CipherExampleResp)(nil), // 4: api.v1.hello.CipherExampleResp
	(*SayReq)(nil),            // 5: api.v1.hello.SayReq
	(*Configs)(nil),           // 6: api.v1.hello.Configs
	(*SayObj)(nil),            // 7: api.v1.hello.SayObj
	(*emptypb.Empty)(nil),     // 8: google.protobuf.Empty
}
var file_hello_hello_proto_depIdxs = []int32{
	7,  // 0: api.v1.hello.SayReq.obj:type_name -> api.v1.hello.SayObj
	7,  // 1: api.v1.hello.SayReq.objs:type_name -> api.v1.hello.SayObj
	6,  // 2: api.v1.hello.SayReq.configs:type_name -> api.v1.hello.Configs
	0,  // 3: api.v1.hello.SayReq.kind:type_name -> api.v1.hello.Kind
	0,  // 4: api.v1.hello.SayReq.kinds:type_name -> api.v1.hello.Kind
	5,  // 5: api.v1.hello.Hello.Say:input_type -> api.v1.hello.SayReq
	5,  // 6: api.v1.hello.Hello.Call:input_type -> api.v1.hello.SayReq
	8,  // 7: api.v1.hello.Hello.Log:input_type -> google.protobuf.Empty
	1,  // 8: api.v1.hello.Hello.CipherExample:input_type -> api.v1.hello.CipherExampleReq
	2,  // 9: api.v1.hello.Hello.MysqlExample:input_type -> api.v1.hello.MysqlExampleReq
	5,  // 10: api.v1.hello.Hello.Say:output_type -> api.v1.hello.SayReq
	5,  // 11: api.v1.hello.Hello.Call:output_type -> api.v1.hello.SayReq
	8,  // 12: api.v1.hello.Hello.Log:output_type -> google.protobuf.Empty
	4,  // 13: api.v1.hello.Hello.CipherExample:output_type -> api.v1.hello.CipherExampleResp
	3,  // 14: api.v1.hello.Hello.MysqlExample:output_type -> api.v1.hello.MysqlExampleResp
	10, // [10:15] is the sub-list for method output_type
	5,  // [5:10] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
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
			switch v := v.(*Configs); i {
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
		file_hello_hello_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
	file_hello_hello_proto_msgTypes[4].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_hello_hello_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hello_hello_proto_goTypes,
		DependencyIndexes: file_hello_hello_proto_depIdxs,
		EnumInfos:         file_hello_hello_proto_enumTypes,
		MessageInfos:      file_hello_hello_proto_msgTypes,
	}.Build()
	File_hello_hello_proto = out.File
	file_hello_hello_proto_rawDesc = nil
	file_hello_hello_proto_goTypes = nil
	file_hello_hello_proto_depIdxs = nil
}
