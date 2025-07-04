// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: mq.proto

package mqpb

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

type MQ struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 交换机
	//
	// Types that are assignable to Exchange:
	//
	//	*MQ_Direct
	//	*MQ_Topic
	//	*MQ_Fanout
	//	*MQ_Headers
	Exchange isMQ_Exchange `protobuf_oneof:"exchange"`
	// 路由key
	Route string `protobuf:"bytes,10,opt,name=route,proto3" json:"route,omitempty"`
	// 消费者
	Consumer string `protobuf:"bytes,11,opt,name=consumer,proto3" json:"consumer,omitempty"`
	// 自动ack
	AutoAck *bool `protobuf:"varint,12,opt,name=auto_ack,json=autoAck,proto3,oneof" json:"auto_ack,omitempty"`
	// 是否持久化
	Durable *bool `protobuf:"varint,13,opt,name=durable,proto3,oneof" json:"durable,omitempty"`
	// 是否自动删除
	AutoDelete *bool `protobuf:"varint,14,opt,name=auto_delete,json=autoDelete,proto3,oneof" json:"auto_delete,omitempty"`
	// 是否排他
	Exclusive *bool `protobuf:"varint,15,opt,name=exclusive,proto3,oneof" json:"exclusive,omitempty"`
	// 设置为true，表示 不能将同一个Conenction中生产者发送的消息传递给这个Connection中 的消费者
	NoLocal *bool `protobuf:"varint,16,opt,name=no_local,json=noLocal,proto3,oneof" json:"no_local,omitempty"`
	// 是否阻塞
	NoWait   *bool             `protobuf:"varint,17,opt,name=no_wait,json=noWait,proto3,oneof" json:"no_wait,omitempty"`
	Internal *bool             `protobuf:"varint,18,opt,name=internal,proto3,oneof" json:"internal,omitempty"`
	Table    map[string]string `protobuf:"bytes,19,rep,name=table,proto3" json:"table,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MQ) Reset() {
	*x = MQ{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mq_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MQ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MQ) ProtoMessage() {}

func (x *MQ) ProtoReflect() protoreflect.Message {
	mi := &file_mq_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MQ.ProtoReflect.Descriptor instead.
func (*MQ) Descriptor() ([]byte, []int) {
	return file_mq_proto_rawDescGZIP(), []int{0}
}

func (m *MQ) GetExchange() isMQ_Exchange {
	if m != nil {
		return m.Exchange
	}
	return nil
}

func (x *MQ) GetDirect() string {
	if x, ok := x.GetExchange().(*MQ_Direct); ok {
		return x.Direct
	}
	return ""
}

func (x *MQ) GetTopic() string {
	if x, ok := x.GetExchange().(*MQ_Topic); ok {
		return x.Topic
	}
	return ""
}

func (x *MQ) GetFanout() string {
	if x, ok := x.GetExchange().(*MQ_Fanout); ok {
		return x.Fanout
	}
	return ""
}

func (x *MQ) GetHeaders() string {
	if x, ok := x.GetExchange().(*MQ_Headers); ok {
		return x.Headers
	}
	return ""
}

func (x *MQ) GetRoute() string {
	if x != nil {
		return x.Route
	}
	return ""
}

func (x *MQ) GetConsumer() string {
	if x != nil {
		return x.Consumer
	}
	return ""
}

func (x *MQ) GetAutoAck() bool {
	if x != nil && x.AutoAck != nil {
		return *x.AutoAck
	}
	return false
}

func (x *MQ) GetDurable() bool {
	if x != nil && x.Durable != nil {
		return *x.Durable
	}
	return false
}

func (x *MQ) GetAutoDelete() bool {
	if x != nil && x.AutoDelete != nil {
		return *x.AutoDelete
	}
	return false
}

func (x *MQ) GetExclusive() bool {
	if x != nil && x.Exclusive != nil {
		return *x.Exclusive
	}
	return false
}

func (x *MQ) GetNoLocal() bool {
	if x != nil && x.NoLocal != nil {
		return *x.NoLocal
	}
	return false
}

func (x *MQ) GetNoWait() bool {
	if x != nil && x.NoWait != nil {
		return *x.NoWait
	}
	return false
}

func (x *MQ) GetInternal() bool {
	if x != nil && x.Internal != nil {
		return *x.Internal
	}
	return false
}

func (x *MQ) GetTable() map[string]string {
	if x != nil {
		return x.Table
	}
	return nil
}

type isMQ_Exchange interface {
	isMQ_Exchange()
}

type MQ_Direct struct {
	Direct string `protobuf:"bytes,1,opt,name=direct,proto3,oneof"`
}

type MQ_Topic struct {
	Topic string `protobuf:"bytes,2,opt,name=topic,proto3,oneof"`
}

type MQ_Fanout struct {
	Fanout string `protobuf:"bytes,3,opt,name=fanout,proto3,oneof"`
}

type MQ_Headers struct {
	Headers string `protobuf:"bytes,4,opt,name=headers,proto3,oneof"`
}

func (*MQ_Direct) isMQ_Exchange() {}

func (*MQ_Topic) isMQ_Exchange() {}

func (*MQ_Fanout) isMQ_Exchange() {}

func (*MQ_Headers) isMQ_Exchange() {}

var file_mq_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: ([]*MQ)(nil),
		Field:         70000,
		Name:          "asjard.api.mq",
		Tag:           "bytes,70000,rep,name=mq",
		Filename:      "mq.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*MQ)(nil),
		Field:         80000,
		Name:          "asjard.api.serviceMq",
		Tag:           "bytes,80000,opt,name=serviceMq",
		Filename:      "mq.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// repeated asjard.api.MQ mq = 70000;
	E_Mq = &file_mq_proto_extTypes[0]
)

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional asjard.api.MQ serviceMq = 80000;
	E_ServiceMq = &file_mq_proto_extTypes[1]
)

var File_mq_proto protoreflect.FileDescriptor

var file_mq_proto_rawDesc = []byte{
	0x0a, 0x08, 0x6d, 0x71, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x61, 0x73, 0x6a, 0x61,
	0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x04, 0x0a, 0x02, 0x4d, 0x51, 0x12,
	0x18, 0x0a, 0x06, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x00, 0x52, 0x06, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x05, 0x74, 0x6f, 0x70,
	0x69, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69,
	0x63, 0x12, 0x18, 0x0a, 0x06, 0x66, 0x61, 0x6e, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x06, 0x66, 0x61, 0x6e, 0x6f, 0x75, 0x74, 0x12, 0x1a, 0x0a, 0x07, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07,
	0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x08, 0x61, 0x75, 0x74,
	0x6f, 0x5f, 0x61, 0x63, 0x6b, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x48, 0x01, 0x52, 0x07, 0x61,
	0x75, 0x74, 0x6f, 0x41, 0x63, 0x6b, 0x88, 0x01, 0x01, 0x12, 0x1d, 0x0a, 0x07, 0x64, 0x75, 0x72,
	0x61, 0x62, 0x6c, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x08, 0x48, 0x02, 0x52, 0x07, 0x64, 0x75,
	0x72, 0x61, 0x62, 0x6c, 0x65, 0x88, 0x01, 0x01, 0x12, 0x24, 0x0a, 0x0b, 0x61, 0x75, 0x74, 0x6f,
	0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x08, 0x48, 0x03, 0x52,
	0x0a, 0x61, 0x75, 0x74, 0x6f, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x88, 0x01, 0x01, 0x12, 0x21,
	0x0a, 0x09, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x73, 0x69, 0x76, 0x65, 0x18, 0x0f, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x04, 0x52, 0x09, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x73, 0x69, 0x76, 0x65, 0x88, 0x01,
	0x01, 0x12, 0x1e, 0x0a, 0x08, 0x6e, 0x6f, 0x5f, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x18, 0x10, 0x20,
	0x01, 0x28, 0x08, 0x48, 0x05, 0x52, 0x07, 0x6e, 0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x88, 0x01,
	0x01, 0x12, 0x1c, 0x0a, 0x07, 0x6e, 0x6f, 0x5f, 0x77, 0x61, 0x69, 0x74, 0x18, 0x11, 0x20, 0x01,
	0x28, 0x08, 0x48, 0x06, 0x52, 0x06, 0x6e, 0x6f, 0x57, 0x61, 0x69, 0x74, 0x88, 0x01, 0x01, 0x12,
	0x1f, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x18, 0x12, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x07, 0x52, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x88, 0x01, 0x01,
	0x12, 0x2f, 0x0a, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x13, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4d, 0x51, 0x2e,
	0x54, 0x61, 0x62, 0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x74, 0x61, 0x62, 0x6c,
	0x65, 0x1a, 0x38, 0x0a, 0x0a, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x65,
	0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x61, 0x75, 0x74, 0x6f,
	0x5f, 0x61, 0x63, 0x6b, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x64, 0x75, 0x72, 0x61, 0x62, 0x6c, 0x65,
	0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x73, 0x69, 0x76, 0x65, 0x42, 0x0b,
	0x0a, 0x09, 0x5f, 0x6e, 0x6f, 0x5f, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x42, 0x0a, 0x0a, 0x08, 0x5f,
	0x6e, 0x6f, 0x5f, 0x77, 0x61, 0x69, 0x74, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x3a, 0x40, 0x0a, 0x02, 0x6d, 0x71, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf0, 0xa2, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x4d, 0x51, 0x52, 0x02, 0x6d, 0x71, 0x3a, 0x4f, 0x0a, 0x09, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x4d, 0x71, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x80, 0xf1, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61,
	0x73, 0x6a, 0x61, 0x72, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4d, 0x51, 0x52, 0x09, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x71, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x73, 0x6a, 0x61, 0x72, 0x64, 0x2f, 0x61, 0x73, 0x6a,
	0x61, 0x72, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x6d, 0x71, 0x70, 0x62, 0x3b, 0x6d, 0x71, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_mq_proto_rawDescOnce sync.Once
	file_mq_proto_rawDescData = file_mq_proto_rawDesc
)

func file_mq_proto_rawDescGZIP() []byte {
	file_mq_proto_rawDescOnce.Do(func() {
		file_mq_proto_rawDescData = protoimpl.X.CompressGZIP(file_mq_proto_rawDescData)
	})
	return file_mq_proto_rawDescData
}

var file_mq_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mq_proto_goTypes = []interface{}{
	(*MQ)(nil),                          // 0: asjard.api.MQ
	nil,                                 // 1: asjard.api.MQ.TableEntry
	(*descriptorpb.MethodOptions)(nil),  // 2: google.protobuf.MethodOptions
	(*descriptorpb.ServiceOptions)(nil), // 3: google.protobuf.ServiceOptions
}
var file_mq_proto_depIdxs = []int32{
	1, // 0: asjard.api.MQ.table:type_name -> asjard.api.MQ.TableEntry
	2, // 1: asjard.api.mq:extendee -> google.protobuf.MethodOptions
	3, // 2: asjard.api.serviceMq:extendee -> google.protobuf.ServiceOptions
	0, // 3: asjard.api.mq:type_name -> asjard.api.MQ
	0, // 4: asjard.api.serviceMq:type_name -> asjard.api.MQ
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	3, // [3:5] is the sub-list for extension type_name
	1, // [1:3] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_mq_proto_init() }
func file_mq_proto_init() {
	if File_mq_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mq_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MQ); i {
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
	file_mq_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*MQ_Direct)(nil),
		(*MQ_Topic)(nil),
		(*MQ_Fanout)(nil),
		(*MQ_Headers)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mq_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_mq_proto_goTypes,
		DependencyIndexes: file_mq_proto_depIdxs,
		MessageInfos:      file_mq_proto_msgTypes,
		ExtensionInfos:    file_mq_proto_extTypes,
	}.Build()
	File_mq_proto = out.File
	file_mq_proto_rawDesc = nil
	file_mq_proto_goTypes = nil
	file_mq_proto_depIdxs = nil
}
