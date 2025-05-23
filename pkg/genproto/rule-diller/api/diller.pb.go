// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v5.29.3
// source: rule-diller/api/diller.proto

package rulediller

import (
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

type GetRuleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service string `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	Context string `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
}

func (x *GetRuleRequest) Reset() {
	*x = GetRuleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rule_diller_api_diller_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRuleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRuleRequest) ProtoMessage() {}

func (x *GetRuleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rule_diller_api_diller_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRuleRequest.ProtoReflect.Descriptor instead.
func (*GetRuleRequest) Descriptor() ([]byte, []int) {
	return file_rule_diller_api_diller_proto_rawDescGZIP(), []int{0}
}

func (x *GetRuleRequest) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *GetRuleRequest) GetContext() string {
	if x != nil {
		return x.Context
	}
	return ""
}

type GetRuleDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data    string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Payload string `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *GetRuleDataResponse) Reset() {
	*x = GetRuleDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rule_diller_api_diller_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRuleDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRuleDataResponse) ProtoMessage() {}

func (x *GetRuleDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rule_diller_api_diller_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRuleDataResponse.ProtoReflect.Descriptor instead.
func (*GetRuleDataResponse) Descriptor() ([]byte, []int) {
	return file_rule_diller_api_diller_proto_rawDescGZIP(), []int{1}
}

func (x *GetRuleDataResponse) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

func (x *GetRuleDataResponse) GetPayload() string {
	if x != nil {
		return x.Payload
	}
	return ""
}

type GetRuleStatisticResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Scores []*VariantScore `protobuf:"bytes,1,rep,name=scores,proto3" json:"scores,omitempty"`
}

func (x *GetRuleStatisticResponse) Reset() {
	*x = GetRuleStatisticResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rule_diller_api_diller_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRuleStatisticResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRuleStatisticResponse) ProtoMessage() {}

func (x *GetRuleStatisticResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rule_diller_api_diller_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRuleStatisticResponse.ProtoReflect.Descriptor instead.
func (*GetRuleStatisticResponse) Descriptor() ([]byte, []int) {
	return file_rule_diller_api_diller_proto_rawDescGZIP(), []int{2}
}

func (x *GetRuleStatisticResponse) GetScores() []*VariantScore {
	if x != nil {
		return x.Scores
	}
	return nil
}

type VariantScore struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VariantId string  `protobuf:"bytes,1,opt,name=variant_id,json=variantId,proto3" json:"variant_id,omitempty"`
	Score     float64 `protobuf:"fixed64,2,opt,name=score,proto3" json:"score,omitempty"`
}

func (x *VariantScore) Reset() {
	*x = VariantScore{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rule_diller_api_diller_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VariantScore) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VariantScore) ProtoMessage() {}

func (x *VariantScore) ProtoReflect() protoreflect.Message {
	mi := &file_rule_diller_api_diller_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VariantScore.ProtoReflect.Descriptor instead.
func (*VariantScore) Descriptor() ([]byte, []int) {
	return file_rule_diller_api_diller_proto_rawDescGZIP(), []int{3}
}

func (x *VariantScore) GetVariantId() string {
	if x != nil {
		return x.VariantId
	}
	return ""
}

func (x *VariantScore) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

var File_rule_diller_api_diller_proto protoreflect.FileDescriptor

var file_rule_diller_api_diller_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x72, 0x75, 0x6c, 0x65, 0x2d, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a,
	0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e,
	0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x44, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x52,
	0x75, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x22, 0x43,
	0x0a, 0x13, 0x47, 0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x5c, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x40, 0x0a, 0x06, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x28, 0x2e, 0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x56, 0x61, 0x72,
	0x69, 0x61, 0x6e, 0x74, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x52, 0x06, 0x73, 0x63, 0x6f, 0x72, 0x65,
	0x73, 0x22, 0x43, 0x0a, 0x0c, 0x56, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x53, 0x63, 0x6f, 0x72,
	0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x49, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x32, 0xb8, 0x02, 0x0a, 0x11, 0x52, 0x75, 0x6c, 0x65, 0x44,
	0x69, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x97, 0x01, 0x0a,
	0x10, 0x47, 0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69,
	0x63, 0x12, 0x2a, 0x2e, 0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e,
	0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e,
	0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x75,
	0x6c, 0x65, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x12, 0x19, 0x2f, 0x76, 0x31,
	0x2f, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x72, 0x75, 0x6c, 0x65, 0x2f, 0x73, 0x74, 0x61,
	0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x12, 0x88, 0x01, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x75,
	0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x2a, 0x2e, 0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2f, 0x2e, 0x62, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2e,
	0x47, 0x65, 0x74, 0x52, 0x75, 0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x12, 0x14, 0x2f, 0x76, 0x31,
	0x2f, 0x64, 0x69, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x72, 0x75, 0x6c, 0x65, 0x2f, 0x64, 0x61, 0x74,
	0x61, 0x42, 0x0e, 0x5a, 0x0c, 0x2e, 0x2f, 0x72, 0x75, 0x6c, 0x65, 0x64, 0x69, 0x6c, 0x6c, 0x65,
	0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rule_diller_api_diller_proto_rawDescOnce sync.Once
	file_rule_diller_api_diller_proto_rawDescData = file_rule_diller_api_diller_proto_rawDesc
)

func file_rule_diller_api_diller_proto_rawDescGZIP() []byte {
	file_rule_diller_api_diller_proto_rawDescOnce.Do(func() {
		file_rule_diller_api_diller_proto_rawDescData = protoimpl.X.CompressGZIP(file_rule_diller_api_diller_proto_rawDescData)
	})
	return file_rule_diller_api_diller_proto_rawDescData
}

var file_rule_diller_api_diller_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_rule_diller_api_diller_proto_goTypes = []interface{}{
	(*GetRuleRequest)(nil),           // 0: bandit.services.rulediller.GetRuleRequest
	(*GetRuleDataResponse)(nil),      // 1: bandit.services.rulediller.GetRuleDataResponse
	(*GetRuleStatisticResponse)(nil), // 2: bandit.services.rulediller.GetRuleStatisticResponse
	(*VariantScore)(nil),             // 3: bandit.services.rulediller.VariantScore
}
var file_rule_diller_api_diller_proto_depIdxs = []int32{
	3, // 0: bandit.services.rulediller.GetRuleStatisticResponse.scores:type_name -> bandit.services.rulediller.VariantScore
	0, // 1: bandit.services.rulediller.RuleDillerService.GetRuleStatistic:input_type -> bandit.services.rulediller.GetRuleRequest
	0, // 2: bandit.services.rulediller.RuleDillerService.GetRuleData:input_type -> bandit.services.rulediller.GetRuleRequest
	2, // 3: bandit.services.rulediller.RuleDillerService.GetRuleStatistic:output_type -> bandit.services.rulediller.GetRuleStatisticResponse
	1, // 4: bandit.services.rulediller.RuleDillerService.GetRuleData:output_type -> bandit.services.rulediller.GetRuleDataResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rule_diller_api_diller_proto_init() }
func file_rule_diller_api_diller_proto_init() {
	if File_rule_diller_api_diller_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rule_diller_api_diller_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRuleRequest); i {
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
		file_rule_diller_api_diller_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRuleDataResponse); i {
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
		file_rule_diller_api_diller_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRuleStatisticResponse); i {
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
		file_rule_diller_api_diller_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VariantScore); i {
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
			RawDescriptor: file_rule_diller_api_diller_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rule_diller_api_diller_proto_goTypes,
		DependencyIndexes: file_rule_diller_api_diller_proto_depIdxs,
		MessageInfos:      file_rule_diller_api_diller_proto_msgTypes,
	}.Build()
	File_rule_diller_api_diller_proto = out.File
	file_rule_diller_api_diller_proto_rawDesc = nil
	file_rule_diller_api_diller_proto_goTypes = nil
	file_rule_diller_api_diller_proto_depIdxs = nil
}
