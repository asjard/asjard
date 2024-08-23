// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: http.proto

package httppb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Http struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Refer to [selector][google.api.DocumentationRule.selector] for syntax details.
	// Determines the URL pattern is matched by this rules. This pattern can be
	// used with any of the {get|put|post|delete|patch} methods. A custom method
	// can be defined using the 'custom' field.
	//
	// Types that are assignable to Pattern:
	//
	//	*Http_Get
	//	*Http_Put
	//	*Http_Post
	//	*Http_Delete
	//	*Http_Patch
	//	*Http_Head
	//	*Http_Options
	Pattern isHttp_Pattern `protobuf_oneof:"pattern"`
	// api类型，如果为空则从package中解析
	Api string `protobuf:"bytes,9,opt,name=api,proto3" json:"api,omitempty"`
	// 接口版本，如果为空则从package中解析
	Version string `protobuf:"bytes,10,opt,name=version,proto3" json:"version,omitempty"`
	// 接口分组
	Group string `protobuf:"bytes,11,opt,name=group,proto3" json:"group,omitempty"`
	// rest writer名称
	WriterName string `protobuf:"bytes,12,opt,name=writer_name,json=writerName,proto3" json:"writer_name,omitempty"`
}

func (x *Http) Reset() {
	*x = Http{}
	if protoimpl.UnsafeEnabled {
		mi := &file_http_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Http) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Http) ProtoMessage() {}

func (x *Http) ProtoReflect() protoreflect.Message {
	mi := &file_http_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Http.ProtoReflect.Descriptor instead.
func (*Http) Descriptor() ([]byte, []int) {
	return file_http_proto_rawDescGZIP(), []int{0}
}

func (m *Http) GetPattern() isHttp_Pattern {
	if m != nil {
		return m.Pattern
	}
	return nil
}

func (x *Http) GetGet() string {
	if x, ok := x.GetPattern().(*Http_Get); ok {
		return x.Get
	}
	return ""
}

func (x *Http) GetPut() string {
	if x, ok := x.GetPattern().(*Http_Put); ok {
		return x.Put
	}
	return ""
}

func (x *Http) GetPost() string {
	if x, ok := x.GetPattern().(*Http_Post); ok {
		return x.Post
	}
	return ""
}

func (x *Http) GetDelete() string {
	if x, ok := x.GetPattern().(*Http_Delete); ok {
		return x.Delete
	}
	return ""
}

func (x *Http) GetPatch() string {
	if x, ok := x.GetPattern().(*Http_Patch); ok {
		return x.Patch
	}
	return ""
}

func (x *Http) GetHead() string {
	if x, ok := x.GetPattern().(*Http_Head); ok {
		return x.Head
	}
	return ""
}

func (x *Http) GetOptions() string {
	if x, ok := x.GetPattern().(*Http_Options); ok {
		return x.Options
	}
	return ""
}

func (x *Http) GetApi() string {
	if x != nil {
		return x.Api
	}
	return ""
}

func (x *Http) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Http) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

func (x *Http) GetWriterName() string {
	if x != nil {
		return x.WriterName
	}
	return ""
}

type isHttp_Pattern interface {
	isHttp_Pattern()
}

type Http_Get struct {
	// Used for listing and getting information about resources.
	Get string `protobuf:"bytes,2,opt,name=get,proto3,oneof"`
}

type Http_Put struct {
	// Used for updating a resource.
	Put string `protobuf:"bytes,3,opt,name=put,proto3,oneof"`
}

type Http_Post struct {
	// Used for creating a resource.
	Post string `protobuf:"bytes,4,opt,name=post,proto3,oneof"`
}

type Http_Delete struct {
	// Used for deleting a resource.
	Delete string `protobuf:"bytes,5,opt,name=delete,proto3,oneof"`
}

type Http_Patch struct {
	// Used for updating a resource.
	Patch string `protobuf:"bytes,6,opt,name=patch,proto3,oneof"`
}

type Http_Head struct {
	// Used for check a resource
	Head string `protobuf:"bytes,7,opt,name=head,proto3,oneof"`
}

type Http_Options struct {
	Options string `protobuf:"bytes,8,opt,name=options,proto3,oneof"`
}

func (*Http_Get) isHttp_Pattern() {}

func (*Http_Put) isHttp_Pattern() {}

func (*Http_Post) isHttp_Pattern() {}

func (*Http_Delete) isHttp_Pattern() {}

func (*Http_Patch) isHttp_Pattern() {}

func (*Http_Head) isHttp_Pattern() {}

func (*Http_Options) isHttp_Pattern() {}

var file_http_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: ([]*Http)(nil),
		Field:         50000,
		Name:          "asjard.api.http",
		Tag:           "bytes,50000,rep,name=http",
		Filename:      "http.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*Http)(nil),
		Field:         60000,
		Name:          "asjard.api.serviceHttp",
		Tag:           "bytes,60000,opt,name=serviceHttp",
		Filename:      "http.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// repeated asjard.api.Http http = 50000;
	E_Http = &file_http_proto_extTypes[0]
)

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional asjard.api.Http serviceHttp = 60000;
	E_ServiceHttp = &file_http_proto_extTypes[1]
)

var File_http_proto protoreflect.FileDescriptor

var file_http_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x61, 0x73,
	0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x96, 0x02, 0x0a, 0x04, 0x48,
	0x74, 0x74, 0x70, 0x12, 0x12, 0x0a, 0x03, 0x67, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x03, 0x67, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x03, 0x70, 0x75, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x70, 0x75, 0x74, 0x12, 0x14, 0x0a, 0x04, 0x70,
	0x6f, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x70, 0x6f, 0x73,
	0x74, 0x12, 0x18, 0x0a, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x05, 0x70,
	0x61, 0x74, 0x63, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x70, 0x61,
	0x74, 0x63, 0x68, 0x12, 0x14, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x1a, 0x0a, 0x07, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x70, 0x69, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x61, 0x70, 0x69, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x14, 0x0a, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x72, 0x69, 0x74, 0x65,
	0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x72,
	0x69, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x3a, 0x46, 0x0a, 0x04, 0x68, 0x74, 0x74, 0x70, 0x12, 0x1e, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x86, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x48, 0x74, 0x74, 0x70, 0x52, 0x04, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x55, 0x0a, 0x0b, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x48, 0x74, 0x74, 0x70, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe0, 0xd4, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x48, 0x74, 0x74, 0x70, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x48, 0x74,
	0x74, 0x70, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x68, 0x74, 0x74, 0x70,
	0x70, 0x62, 0x3b, 0x68, 0x74, 0x74, 0x70, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_http_proto_rawDescOnce sync.Once
	file_http_proto_rawDescData = file_http_proto_rawDesc
)

func file_http_proto_rawDescGZIP() []byte {
	file_http_proto_rawDescOnce.Do(func() {
		file_http_proto_rawDescData = protoimpl.X.CompressGZIP(file_http_proto_rawDescData)
	})
	return file_http_proto_rawDescData
}

var file_http_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_http_proto_goTypes = []interface{}{
	(*Http)(nil),                        // 0: asjard.api.Http
	(*descriptorpb.MethodOptions)(nil),  // 1: google.protobuf.MethodOptions
	(*descriptorpb.ServiceOptions)(nil), // 2: google.protobuf.ServiceOptions
}
var file_http_proto_depIdxs = []int32{
	1, // 0: asjard.api.http:extendee -> google.protobuf.MethodOptions
	2, // 1: asjard.api.serviceHttp:extendee -> google.protobuf.ServiceOptions
	0, // 2: asjard.api.http:type_name -> asjard.api.Http
	0, // 3: asjard.api.serviceHttp:type_name -> asjard.api.Http
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	2, // [2:4] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_http_proto_init() }
func file_http_proto_init() {
	if File_http_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_http_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Http); i {
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
	file_http_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Http_Get)(nil),
		(*Http_Put)(nil),
		(*Http_Post)(nil),
		(*Http_Delete)(nil),
		(*Http_Patch)(nil),
		(*Http_Head)(nil),
		(*Http_Options)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_http_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_http_proto_goTypes,
		DependencyIndexes: file_http_proto_depIdxs,
		MessageInfos:      file_http_proto_msgTypes,
		ExtensionInfos:    file_http_proto_extTypes,
	}.Build()
	File_http_proto = out.File
	file_http_proto_rawDesc = nil
	file_http_proto_goTypes = nil
	file_http_proto_depIdxs = nil
}
