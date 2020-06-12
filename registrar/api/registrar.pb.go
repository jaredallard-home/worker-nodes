// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.11.4
// source: registrar.proto

package api

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RegisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID is used to correlate devices. This will be
	// returned from an initial register request.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// authToken allows access to this endpoint
	AuthToken string `protobuf:"bytes,2,opt,name=authToken,proto3" json:"authToken,omitempty"`
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_registrar_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_registrar_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_registrar_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RegisterRequest) GetAuthToken() string {
	if x != nil {
		return x.AuthToken
	}
	return ""
}

type RegisterResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID becomes this device's unique ID
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Key is the wireguard key that has been generated for this device.
	Key string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	// RancherToken is the token used to register the node with Rancher.
	RancherToken string `protobuf:"bytes,3,opt,name=rancher_token,json=rancherToken,proto3" json:"rancher_token,omitempty"`
	// IPAddress is the IP that this node should be running as
	IpAddress string `protobuf:"bytes,4,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	// PublicKey is the server public key
	PublicKey string `protobuf:"bytes,5,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *RegisterResponse) Reset() {
	*x = RegisterResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_registrar_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterResponse) ProtoMessage() {}

func (x *RegisterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_registrar_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterResponse.ProtoReflect.Descriptor instead.
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return file_registrar_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RegisterResponse) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *RegisterResponse) GetRancherToken() string {
	if x != nil {
		return x.RancherToken
	}
	return ""
}

func (x *RegisterResponse) GetIpAddress() string {
	if x != nil {
		return x.IpAddress
	}
	return ""
}

func (x *RegisterResponse) GetPublicKey() string {
	if x != nil {
		return x.PublicKey
	}
	return ""
}

var File_registrar_proto protoreflect.FileDescriptor

var file_registrar_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x61, 0x70, 0x69, 0x22, 0x3f, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x75, 0x74,
	0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x75,
	0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x97, 0x01, 0x0a, 0x10, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x23,
	0x0a, 0x0d, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x72, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x72, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x32, 0x46, 0x0a, 0x09, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x72, 0x12, 0x39,
	0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x14, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x15, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x65, 0x74, 0x6f, 0x75, 0x74, 0x72, 0x65,
	0x61, 0x63, 0x68, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x7a, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_registrar_proto_rawDescOnce sync.Once
	file_registrar_proto_rawDescData = file_registrar_proto_rawDesc
)

func file_registrar_proto_rawDescGZIP() []byte {
	file_registrar_proto_rawDescOnce.Do(func() {
		file_registrar_proto_rawDescData = protoimpl.X.CompressGZIP(file_registrar_proto_rawDescData)
	})
	return file_registrar_proto_rawDescData
}

var file_registrar_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_registrar_proto_goTypes = []interface{}{
	(*RegisterRequest)(nil),  // 0: api.RegisterRequest
	(*RegisterResponse)(nil), // 1: api.RegisterResponse
}
var file_registrar_proto_depIdxs = []int32{
	0, // 0: api.Registrar.Register:input_type -> api.RegisterRequest
	1, // 1: api.Registrar.Register:output_type -> api.RegisterResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_registrar_proto_init() }
func file_registrar_proto_init() {
	if File_registrar_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_registrar_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterRequest); i {
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
		file_registrar_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterResponse); i {
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
			RawDescriptor: file_registrar_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_registrar_proto_goTypes,
		DependencyIndexes: file_registrar_proto_depIdxs,
		MessageInfos:      file_registrar_proto_msgTypes,
	}.Build()
	File_registrar_proto = out.File
	file_registrar_proto_rawDesc = nil
	file_registrar_proto_goTypes = nil
	file_registrar_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// RegistrarClient is the client API for Registrar service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type RegistrarClient interface {
	// Define your grpc service interface here
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
}

type registrarClient struct {
	cc grpc.ClientConnInterface
}

func NewRegistrarClient(cc grpc.ClientConnInterface) RegistrarClient {
	return &registrarClient{cc}
}

func (c *registrarClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/api.Registrar/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RegistrarServer is the server API for Registrar service.
type RegistrarServer interface {
	// Define your grpc service interface here
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
}

// UnimplementedRegistrarServer can be embedded to have forward compatible implementations.
type UnimplementedRegistrarServer struct {
}

func (*UnimplementedRegistrarServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}

func RegisterRegistrarServer(s *grpc.Server, srv RegistrarServer) {
	s.RegisterService(&_Registrar_serviceDesc, srv)
}

func _Registrar_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegistrarServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.Registrar/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegistrarServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Registrar_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.Registrar",
	HandlerType: (*RegistrarServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Registrar_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "registrar.proto",
}
