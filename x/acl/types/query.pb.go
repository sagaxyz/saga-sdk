// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: saga/acl/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{0}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{1}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type QueryListAdminsRequest struct {
}

func (m *QueryListAdminsRequest) Reset()         { *m = QueryListAdminsRequest{} }
func (m *QueryListAdminsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryListAdminsRequest) ProtoMessage()    {}
func (*QueryListAdminsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{2}
}
func (m *QueryListAdminsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryListAdminsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryListAdminsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryListAdminsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryListAdminsRequest.Merge(m, src)
}
func (m *QueryListAdminsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryListAdminsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryListAdminsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryListAdminsRequest proto.InternalMessageInfo

type QueryListAdminsResponse struct {
	Admins []*Address `protobuf:"bytes,1,rep,name=admins,proto3" json:"admins,omitempty"`
}

func (m *QueryListAdminsResponse) Reset()         { *m = QueryListAdminsResponse{} }
func (m *QueryListAdminsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryListAdminsResponse) ProtoMessage()    {}
func (*QueryListAdminsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{3}
}
func (m *QueryListAdminsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryListAdminsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryListAdminsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryListAdminsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryListAdminsResponse.Merge(m, src)
}
func (m *QueryListAdminsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryListAdminsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryListAdminsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryListAdminsResponse proto.InternalMessageInfo

func (m *QueryListAdminsResponse) GetAdmins() []*Address {
	if m != nil {
		return m.Admins
	}
	return nil
}

type QueryListAllowedRequest struct {
}

func (m *QueryListAllowedRequest) Reset()         { *m = QueryListAllowedRequest{} }
func (m *QueryListAllowedRequest) String() string { return proto.CompactTextString(m) }
func (*QueryListAllowedRequest) ProtoMessage()    {}
func (*QueryListAllowedRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{4}
}
func (m *QueryListAllowedRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryListAllowedRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryListAllowedRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryListAllowedRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryListAllowedRequest.Merge(m, src)
}
func (m *QueryListAllowedRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryListAllowedRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryListAllowedRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryListAllowedRequest proto.InternalMessageInfo

type QueryListAllowedResponse struct {
	Allowed []*Address `protobuf:"bytes,1,rep,name=allowed,proto3" json:"allowed,omitempty"`
}

func (m *QueryListAllowedResponse) Reset()         { *m = QueryListAllowedResponse{} }
func (m *QueryListAllowedResponse) String() string { return proto.CompactTextString(m) }
func (*QueryListAllowedResponse) ProtoMessage()    {}
func (*QueryListAllowedResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0cedc311d1d5d775, []int{5}
}
func (m *QueryListAllowedResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryListAllowedResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryListAllowedResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryListAllowedResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryListAllowedResponse.Merge(m, src)
}
func (m *QueryListAllowedResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryListAllowedResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryListAllowedResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryListAllowedResponse proto.InternalMessageInfo

func (m *QueryListAllowedResponse) GetAllowed() []*Address {
	if m != nil {
		return m.Allowed
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryParamsRequest)(nil), "saga.acl.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "saga.acl.v1.QueryParamsResponse")
	proto.RegisterType((*QueryListAdminsRequest)(nil), "saga.acl.v1.QueryListAdminsRequest")
	proto.RegisterType((*QueryListAdminsResponse)(nil), "saga.acl.v1.QueryListAdminsResponse")
	proto.RegisterType((*QueryListAllowedRequest)(nil), "saga.acl.v1.QueryListAllowedRequest")
	proto.RegisterType((*QueryListAllowedResponse)(nil), "saga.acl.v1.QueryListAllowedResponse")
}

func init() { proto.RegisterFile("saga/acl/v1/query.proto", fileDescriptor_0cedc311d1d5d775) }

var fileDescriptor_0cedc311d1d5d775 = []byte{
	// 411 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x52, 0x41, 0x6f, 0xda, 0x30,
	0x18, 0x4d, 0xd8, 0x96, 0x49, 0xce, 0x61, 0x9b, 0x41, 0x23, 0x44, 0x53, 0x40, 0x19, 0xd3, 0x38,
	0x6c, 0xb1, 0xc2, 0x7e, 0x01, 0x5c, 0x36, 0x4d, 0x3d, 0xb4, 0x1c, 0x7b, 0x33, 0xc4, 0x4a, 0xa3,
	0x86, 0x38, 0xc4, 0x86, 0x42, 0x8f, 0xfd, 0x05, 0x95, 0xfa, 0xa7, 0x38, 0x22, 0x55, 0x95, 0x7a,
	0xaa, 0x2a, 0xe8, 0x0f, 0xa9, 0xb0, 0x0d, 0x0d, 0x4a, 0xa1, 0xa7, 0x44, 0xef, 0x3d, 0xbf, 0xf7,
	0x3d, 0x7f, 0x06, 0x55, 0x86, 0x43, 0x8c, 0xf0, 0x20, 0x46, 0x13, 0x1f, 0x8d, 0xc6, 0x24, 0x9b,
	0x79, 0x69, 0x46, 0x39, 0x85, 0xe6, 0x9a, 0xf0, 0xf0, 0x20, 0xf6, 0x26, 0xbe, 0x5d, 0x09, 0x69,
	0x48, 0x05, 0x8e, 0xd6, 0x7f, 0x52, 0x62, 0x7f, 0x0b, 0x29, 0x0d, 0x63, 0x82, 0x70, 0x1a, 0x21,
	0x9c, 0x24, 0x94, 0x63, 0x1e, 0xd1, 0x84, 0x29, 0xb6, 0x96, 0x77, 0x0e, 0x49, 0x42, 0x58, 0xf4,
	0x2a, 0x85, 0x83, 0x20, 0x23, 0x4c, 0x51, 0x6e, 0x05, 0xc0, 0x93, 0xf5, 0x14, 0xc7, 0x38, 0xc3,
	0x43, 0xd6, 0x23, 0xa3, 0x31, 0x61, 0xdc, 0xfd, 0x07, 0xca, 0x3b, 0x28, 0x4b, 0x69, 0xc2, 0x08,
	0xf4, 0x81, 0x91, 0x0a, 0xc4, 0xd2, 0x1b, 0x7a, 0xcb, 0x6c, 0x97, 0xbd, 0xdc, 0xd0, 0x9e, 0x14,
	0x77, 0xdf, 0xcf, 0x1f, 0xea, 0x5a, 0x4f, 0x09, 0x5d, 0x0b, 0x7c, 0x15, 0x4e, 0x47, 0x11, 0xe3,
	0x9d, 0x60, 0x18, 0x25, 0xdb, 0x8c, 0xbf, 0xa0, 0x5a, 0x60, 0x54, 0xce, 0x2f, 0x60, 0x60, 0x81,
	0x58, 0x7a, 0xe3, 0x5d, 0xcb, 0x6c, 0x57, 0x76, 0x72, 0x3a, 0xb2, 0x40, 0x4f, 0x69, 0xdc, 0x5a,
	0xde, 0x28, 0x8e, 0xe9, 0x05, 0x09, 0x36, 0x19, 0xff, 0x81, 0x55, 0xa4, 0x54, 0x88, 0x07, 0x3e,
	0x62, 0x09, 0x1d, 0x4c, 0xd9, 0x88, 0xda, 0x77, 0x25, 0xf0, 0x41, 0x98, 0xc1, 0x00, 0x18, 0xb2,
	0x2b, 0xac, 0xef, 0x1c, 0x29, 0x5e, 0xa4, 0xdd, 0xd8, 0x2f, 0x90, 0x63, 0xb8, 0xd5, 0xab, 0xdb,
	0xa7, 0x9b, 0xd2, 0x17, 0xf8, 0x09, 0x89, 0x25, 0x4d, 0x7c, 0x24, 0x6f, 0x0e, 0x72, 0x60, 0xe6,
	0xc6, 0x86, 0xcd, 0xa2, 0x53, 0xb1, 0xb0, 0xfd, 0xe3, 0x0d, 0x95, 0x0a, 0xb5, 0x44, 0x28, 0x84,
	0x9f, 0xb7, 0xa1, 0xaa, 0x25, 0x4c, 0x01, 0x78, 0x59, 0x08, 0xfc, 0xbe, 0xc7, 0x2e, 0xbf, 0x48,
	0xbb, 0x79, 0x58, 0xb4, 0xb7, 0xa7, 0x5c, 0x5f, 0xb7, 0x33, 0x5f, 0x3a, 0xfa, 0x62, 0xe9, 0xe8,
	0x8f, 0x4b, 0x47, 0xbf, 0x5e, 0x39, 0xda, 0x62, 0xe5, 0x68, 0xf7, 0x2b, 0x47, 0x3b, 0xfd, 0x19,
	0x46, 0xfc, 0x6c, 0xdc, 0xf7, 0x06, 0x74, 0x28, 0x0e, 0x4d, 0x67, 0x97, 0xe2, 0xfb, 0x9b, 0x05,
	0xe7, 0x68, 0x2a, 0xde, 0x33, 0x9f, 0xa5, 0x84, 0xf5, 0x0d, 0xf1, 0x96, 0xff, 0x3c, 0x07, 0x00,
	0x00, 0xff, 0xff, 0x1b, 0x21, 0x85, 0xcc, 0x5d, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// Params returns the params
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// ListAllowed returns the list of addresses that are allowed to deploy EVM
	// contracts
	ListAllowed(ctx context.Context, in *QueryListAllowedRequest, opts ...grpc.CallOption) (*QueryListAllowedResponse, error)
	// ListAdmins returns the list of admin addresses
	ListAdmins(ctx context.Context, in *QueryListAdminsRequest, opts ...grpc.CallOption) (*QueryListAdminsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/saga.acl.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ListAllowed(ctx context.Context, in *QueryListAllowedRequest, opts ...grpc.CallOption) (*QueryListAllowedResponse, error) {
	out := new(QueryListAllowedResponse)
	err := c.cc.Invoke(ctx, "/saga.acl.v1.Query/ListAllowed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ListAdmins(ctx context.Context, in *QueryListAdminsRequest, opts ...grpc.CallOption) (*QueryListAdminsResponse, error) {
	out := new(QueryListAdminsResponse)
	err := c.cc.Invoke(ctx, "/saga.acl.v1.Query/ListAdmins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Params returns the params
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// ListAllowed returns the list of addresses that are allowed to deploy EVM
	// contracts
	ListAllowed(context.Context, *QueryListAllowedRequest) (*QueryListAllowedResponse, error)
	// ListAdmins returns the list of admin addresses
	ListAdmins(context.Context, *QueryListAdminsRequest) (*QueryListAdminsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (*UnimplementedQueryServer) ListAllowed(ctx context.Context, req *QueryListAllowedRequest) (*QueryListAllowedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAllowed not implemented")
}
func (*UnimplementedQueryServer) ListAdmins(ctx context.Context, req *QueryListAdminsRequest) (*QueryListAdminsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAdmins not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/saga.acl.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ListAllowed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryListAllowedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ListAllowed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/saga.acl.v1.Query/ListAllowed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ListAllowed(ctx, req.(*QueryListAllowedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ListAdmins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryListAdminsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ListAdmins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/saga.acl.v1.Query/ListAdmins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ListAdmins(ctx, req.(*QueryListAdminsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "saga.acl.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "ListAllowed",
			Handler:    _Query_ListAllowed_Handler,
		},
		{
			MethodName: "ListAdmins",
			Handler:    _Query_ListAdmins_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "saga/acl/v1/query.proto",
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryListAdminsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryListAdminsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryListAdminsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryListAdminsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryListAdminsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryListAdminsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Admins) > 0 {
		for iNdEx := len(m.Admins) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Admins[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryListAllowedRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryListAllowedRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryListAllowedRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryListAllowedResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryListAllowedResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryListAllowedResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Allowed) > 0 {
		for iNdEx := len(m.Allowed) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Allowed[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryListAdminsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryListAdminsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Admins) > 0 {
		for _, e := range m.Admins {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryListAllowedRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryListAllowedResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Allowed) > 0 {
		for _, e := range m.Allowed {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryListAdminsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryListAdminsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryListAdminsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryListAdminsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryListAdminsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryListAdminsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Admins", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Admins = append(m.Admins, &Address{})
			if err := m.Admins[len(m.Admins)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryListAllowedRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryListAllowedRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryListAllowedRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryListAllowedResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryListAllowedResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryListAllowedResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Allowed", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Allowed = append(m.Allowed, &Address{})
			if err := m.Allowed[len(m.Allowed)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
