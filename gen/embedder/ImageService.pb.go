// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.1
// 	protoc        v5.29.2
// source: proto/ImageService.proto

package embedder

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

type Image struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Url           string                 `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Image) Reset() {
	*x = Image{}
	mi := &file_proto_ImageService_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Image) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Image) ProtoMessage() {}

func (x *Image) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ImageService_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Image.ProtoReflect.Descriptor instead.
func (*Image) Descriptor() ([]byte, []int) {
	return file_proto_ImageService_proto_rawDescGZIP(), []int{0}
}

func (x *Image) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

type EmbedRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BaseImage     *Image                 `protobuf:"bytes,1,opt,name=base_image,json=baseImage,proto3" json:"base_image,omitempty"`
	Model         string                 `protobuf:"bytes,2,opt,name=model,proto3" json:"model,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EmbedRequest) Reset() {
	*x = EmbedRequest{}
	mi := &file_proto_ImageService_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EmbedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmbedRequest) ProtoMessage() {}

func (x *EmbedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ImageService_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmbedRequest.ProtoReflect.Descriptor instead.
func (*EmbedRequest) Descriptor() ([]byte, []int) {
	return file_proto_ImageService_proto_rawDescGZIP(), []int{1}
}

func (x *EmbedRequest) GetBaseImage() *Image {
	if x != nil {
		return x.BaseImage
	}
	return nil
}

func (x *EmbedRequest) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

type EmbedResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Embedding     []float32              `protobuf:"fixed32,1,rep,packed,name=embedding,proto3" json:"embedding,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EmbedResponse) Reset() {
	*x = EmbedResponse{}
	mi := &file_proto_ImageService_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EmbedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmbedResponse) ProtoMessage() {}

func (x *EmbedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ImageService_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmbedResponse.ProtoReflect.Descriptor instead.
func (*EmbedResponse) Descriptor() ([]byte, []int) {
	return file_proto_ImageService_proto_rawDescGZIP(), []int{2}
}

func (x *EmbedResponse) GetEmbedding() []float32 {
	if x != nil {
		return x.Embedding
	}
	return nil
}

var File_proto_ImageService_proto protoreflect.FileDescriptor

var file_proto_ImageService_proto_rawDesc = []byte{
	0x0a, 0x18, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x45, 0x6d, 0x62, 0x65,
	0x64, 0x64, 0x65, 0x72, 0x22, 0x19, 0x0a, 0x05, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x22,
	0x54, 0x0a, 0x0c, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x2e, 0x0a, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x49,
	0x6d, 0x61, 0x67, 0x65, 0x52, 0x09, 0x62, 0x61, 0x73, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x22, 0x2d, 0x0a, 0x0d, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64,
	0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x02, 0x52, 0x09, 0x65, 0x6d, 0x62, 0x65, 0x64,
	0x64, 0x69, 0x6e, 0x67, 0x32, 0x46, 0x0a, 0x08, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x65, 0x72,
	0x12, 0x3a, 0x0a, 0x05, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x12, 0x16, 0x2e, 0x45, 0x6d, 0x62, 0x65,
	0x64, 0x64, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x17, 0x2e, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x62,
	0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x11, 0x5a, 0x0f,
	0x2e, 0x2e, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_ImageService_proto_rawDescOnce sync.Once
	file_proto_ImageService_proto_rawDescData = file_proto_ImageService_proto_rawDesc
)

func file_proto_ImageService_proto_rawDescGZIP() []byte {
	file_proto_ImageService_proto_rawDescOnce.Do(func() {
		file_proto_ImageService_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_ImageService_proto_rawDescData)
	})
	return file_proto_ImageService_proto_rawDescData
}

var file_proto_ImageService_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proto_ImageService_proto_goTypes = []any{
	(*Image)(nil),         // 0: Embedder.Image
	(*EmbedRequest)(nil),  // 1: Embedder.EmbedRequest
	(*EmbedResponse)(nil), // 2: Embedder.EmbedResponse
}
var file_proto_ImageService_proto_depIdxs = []int32{
	0, // 0: Embedder.EmbedRequest.base_image:type_name -> Embedder.Image
	1, // 1: Embedder.Embedder.Embed:input_type -> Embedder.EmbedRequest
	2, // 2: Embedder.Embedder.Embed:output_type -> Embedder.EmbedResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_ImageService_proto_init() }
func file_proto_ImageService_proto_init() {
	if File_proto_ImageService_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_ImageService_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_ImageService_proto_goTypes,
		DependencyIndexes: file_proto_ImageService_proto_depIdxs,
		MessageInfos:      file_proto_ImageService_proto_msgTypes,
	}.Build()
	File_proto_ImageService_proto = out.File
	file_proto_ImageService_proto_rawDesc = nil
	file_proto_ImageService_proto_goTypes = nil
	file_proto_ImageService_proto_depIdxs = nil
}
