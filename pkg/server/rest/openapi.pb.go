// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: openapi.proto

package rest

import (
	_ "github.com/asjard/asjard/pkg/protobuf/httppb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_openapi_proto protoreflect.FileDescriptor

var file_openapi_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0a, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x1a, 0x25, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32,
	0xdf, 0x03, 0x0a, 0x07, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50, 0x49, 0x12, 0x8a, 0x01, 0x0a, 0x04,
	0x59, 0x61, 0x6d, 0x6c, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x22, 0x52, 0x82, 0xb5, 0x18, 0x24, 0x6a, 0x14, 0x47, 0x65, 0x74, 0x20,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x79, 0x6d, 0x6c, 0x20, 0x66, 0x69, 0x6c, 0x65,
	0x12, 0x0c, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x79, 0x6d, 0x6c, 0x82, 0xb5,
	0x18, 0x26, 0x6a, 0x15, 0x47, 0x65, 0x74, 0x20, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e,
	0x79, 0x61, 0x6d, 0x6c, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x0d, 0x2f, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x12, 0x62, 0x0a, 0x04, 0x4a, 0x73, 0x6f, 0x6e,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x2a, 0x82, 0xb5, 0x18, 0x26, 0x6a, 0x15, 0x47, 0x65, 0x74, 0x20, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x2e, 0x6a, 0x73, 0x6f, 0x6e, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x0d, 0x2f,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x65, 0x0a, 0x04,
	0x50, 0x61, 0x67, 0x65, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x22, 0x2d, 0x82, 0xb5, 0x18, 0x29, 0x6a, 0x18, 0x52, 0x65, 0x64, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x20, 0x74, 0x6f, 0x20, 0x4f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x20,
	0x70, 0x61, 0x67, 0x65, 0x12, 0x0d, 0x2f, 0x70, 0x61, 0x67, 0x65, 0x2f, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x12, 0x6d, 0x0a, 0x0a, 0x53, 0x63, 0x61, 0x6c, 0x61, 0x72, 0x50, 0x61, 0x67,
	0x65, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x2f, 0x82, 0xb5, 0x18, 0x2b, 0x6a, 0x13, 0x53, 0x63, 0x61, 0x6c, 0x61, 0x72, 0x20,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x20, 0x70, 0x61, 0x67, 0x65, 0x12, 0x14, 0x2f, 0x70,
	0x61, 0x67, 0x65, 0x2f, 0x73, 0x63, 0x61, 0x6c, 0x61, 0x72, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x1a, 0x0d, 0x82, 0xa6, 0x1d, 0x09, 0x4a, 0x01, 0x2f, 0x52, 0x01, 0x2f, 0x5a, 0x01,
	0x2f, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_openapi_proto_goTypes = []interface{}{
	(*emptypb.Empty)(nil), // 0: google.protobuf.Empty
}
var file_openapi_proto_depIdxs = []int32{
	0, // 0: asjard.api.OpenAPI.Yaml:input_type -> google.protobuf.Empty
	0, // 1: asjard.api.OpenAPI.Json:input_type -> google.protobuf.Empty
	0, // 2: asjard.api.OpenAPI.Page:input_type -> google.protobuf.Empty
	0, // 3: asjard.api.OpenAPI.ScalarPage:input_type -> google.protobuf.Empty
	0, // 4: asjard.api.OpenAPI.Yaml:output_type -> google.protobuf.Empty
	0, // 5: asjard.api.OpenAPI.Json:output_type -> google.protobuf.Empty
	0, // 6: asjard.api.OpenAPI.Page:output_type -> google.protobuf.Empty
	0, // 7: asjard.api.OpenAPI.ScalarPage:output_type -> google.protobuf.Empty
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_openapi_proto_init() }
func file_openapi_proto_init() {
	if File_openapi_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_openapi_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_openapi_proto_goTypes,
		DependencyIndexes: file_openapi_proto_depIdxs,
	}.Build()
	File_openapi_proto = out.File
	file_openapi_proto_rawDesc = nil
	file_openapi_proto_goTypes = nil
	file_openapi_proto_depIdxs = nil
}
