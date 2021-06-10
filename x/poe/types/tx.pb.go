// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: confio/poe/v1beta1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	types2 "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/staking/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/regen-network/cosmos-proto"
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

// MsgCreateValidator defines a PoE message for creating a new validator.
// Based on the SDK staking.MsgCreateValidator
type MsgCreateValidator struct {
	// Description meta data
	Description types.Description `protobuf:"bytes,1,opt,name=description,proto3" json:"description"`
	// DelegatorAddress is the bech32 address string
	DelegatorAddress string `protobuf:"bytes,4,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty" yaml:"delegator_address"`
	// ValidatorAddress is the bech32 address string with the
	ValidatorAddress string `protobuf:"bytes,5,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty" yaml:"validator_address"`
	// Pubkey public key
	Pubkey *types1.Any `protobuf:"bytes,6,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
	// Value defines the initial staking amount
	Value types2.Coin `protobuf:"bytes,7,opt,name=value,proto3" json:"value"`
}

func (m *MsgCreateValidator) Reset()         { *m = MsgCreateValidator{} }
func (m *MsgCreateValidator) String() string { return proto.CompactTextString(m) }
func (*MsgCreateValidator) ProtoMessage()    {}
func (*MsgCreateValidator) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2f36f4be4f27cf5, []int{0}
}
func (m *MsgCreateValidator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateValidator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateValidator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateValidator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateValidator.Merge(m, src)
}
func (m *MsgCreateValidator) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateValidator) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateValidator.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateValidator proto.InternalMessageInfo

// MsgCreateValidatorResponse defines the Msg/CreateValidator response type.
type MsgCreateValidatorResponse struct {
}

func (m *MsgCreateValidatorResponse) Reset()         { *m = MsgCreateValidatorResponse{} }
func (m *MsgCreateValidatorResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateValidatorResponse) ProtoMessage()    {}
func (*MsgCreateValidatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2f36f4be4f27cf5, []int{1}
}
func (m *MsgCreateValidatorResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateValidatorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateValidatorResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateValidatorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateValidatorResponse.Merge(m, src)
}
func (m *MsgCreateValidatorResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateValidatorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateValidatorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateValidatorResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgCreateValidator)(nil), "confio.poe.v1beta1.MsgCreateValidator")
	proto.RegisterType((*MsgCreateValidatorResponse)(nil), "confio.poe.v1beta1.MsgCreateValidatorResponse")
}

func init() { proto.RegisterFile("confio/poe/v1beta1/tx.proto", fileDescriptor_c2f36f4be4f27cf5) }

var fileDescriptor_c2f36f4be4f27cf5 = []byte{
	// 456 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0x31, 0x6f, 0xd4, 0x30,
	0x14, 0xc7, 0x13, 0xb5, 0x3d, 0x20, 0x1d, 0x28, 0xd1, 0x0d, 0xb9, 0xa3, 0xca, 0x55, 0xa1, 0x42,
	0x5d, 0xb0, 0xd5, 0x22, 0x96, 0x0e, 0x48, 0xbd, 0x22, 0x24, 0x54, 0x55, 0x42, 0x19, 0x18, 0x58,
	0x2a, 0x27, 0x79, 0x35, 0x56, 0x73, 0x79, 0x56, 0xec, 0x9c, 0x9a, 0x8d, 0x91, 0x91, 0x8f, 0xd0,
	0x0f, 0xc1, 0x87, 0xa8, 0x98, 0x3a, 0x32, 0x55, 0xe8, 0x6e, 0x61, 0xe6, 0x13, 0xa0, 0x8b, 0x9d,
	0x14, 0x35, 0x4b, 0xb7, 0xbc, 0xf7, 0xff, 0xfb, 0x67, 0xfb, 0xfd, 0x1d, 0xef, 0x79, 0x8a, 0xc5,
	0xb9, 0x40, 0x2a, 0x11, 0xe8, 0x7c, 0x3f, 0x01, 0xcd, 0xf6, 0xa9, 0xbe, 0x24, 0xb2, 0x44, 0x8d,
	0xbe, 0x6f, 0x44, 0x22, 0x11, 0x88, 0x15, 0xc7, 0x23, 0x8e, 0xc8, 0x73, 0xa0, 0x8d, 0x23, 0xa9,
	0xce, 0x29, 0x2b, 0x6a, 0x63, 0x1f, 0x0f, 0x39, 0x72, 0x6c, 0x3e, 0xe9, 0xea, 0xcb, 0x76, 0x47,
	0x29, 0xaa, 0x19, 0xaa, 0x33, 0x23, 0x98, 0xc2, 0x4a, 0xa1, 0xa9, 0x68, 0xc2, 0xd4, 0xdd, 0xee,
	0x29, 0x8a, 0xc2, 0xea, 0xbb, 0x56, 0x57, 0x9a, 0x5d, 0x88, 0x82, 0x77, 0x16, 0x5b, 0x1b, 0x57,
	0xf4, 0x75, 0xcd, 0xf3, 0x4f, 0x15, 0x3f, 0x2e, 0x81, 0x69, 0xf8, 0xc4, 0x72, 0x91, 0x31, 0x8d,
	0xa5, 0x7f, 0xe2, 0x6d, 0x66, 0xa0, 0xd2, 0x52, 0x48, 0x2d, 0xb0, 0x08, 0xdc, 0x1d, 0x77, 0x6f,
	0xf3, 0xe0, 0x05, 0xb1, 0x07, 0x68, 0x11, 0x16, 0x49, 0xde, 0xdd, 0x59, 0xa7, 0xeb, 0xd7, 0xb7,
	0x13, 0x27, 0xfe, 0x7f, 0xb5, 0xff, 0xc1, 0x7b, 0x96, 0x41, 0x0e, 0x7c, 0x45, 0x3e, 0x63, 0x59,
	0x56, 0x82, 0x52, 0xc1, 0xfa, 0x8e, 0xbb, 0xf7, 0x64, 0xba, 0xfd, 0xf7, 0x76, 0x12, 0xd4, 0x6c,
	0x96, 0x1f, 0x46, 0x3d, 0x4b, 0x14, 0x6f, 0x75, 0xbd, 0x23, 0xd3, 0x5a, 0xa1, 0xe6, 0xed, 0x21,
	0x3b, 0xd4, 0xc6, 0x7d, 0x54, 0xcf, 0x12, 0xc5, 0x5b, 0x5d, 0xaf, 0x45, 0xbd, 0xf7, 0x06, 0xb2,
	0x4a, 0x2e, 0xa0, 0x0e, 0x06, 0xcd, 0xed, 0x86, 0xc4, 0x84, 0x43, 0xda, 0x70, 0xc8, 0x51, 0x51,
	0x4f, 0x83, 0x9f, 0x3f, 0x5e, 0x0d, 0xed, 0xb5, 0xd3, 0xb2, 0x96, 0x1a, 0xc9, 0xc7, 0x2a, 0x39,
	0x81, 0x3a, 0xb6, 0xab, 0xfd, 0x37, 0xde, 0xc6, 0x9c, 0xe5, 0x15, 0x04, 0x8f, 0x1a, 0xcc, 0xa8,
	0x1d, 0xd2, 0x2a, 0x97, 0x6e, 0x42, 0xc7, 0x28, 0xda, 0xd1, 0x18, 0xf7, 0xe1, 0xe3, 0x6f, 0x57,
	0x13, 0xe7, 0xcf, 0xd5, 0xc4, 0x89, 0xb6, 0xbd, 0x71, 0x3f, 0x81, 0x18, 0x94, 0xc4, 0x42, 0xc1,
	0x81, 0xf4, 0xd6, 0x4e, 0x15, 0xf7, 0x85, 0xf7, 0xf4, 0x7e, 0x46, 0x2f, 0x49, 0xff, 0x85, 0x91,
	0x3e, 0x69, 0x4c, 0x1e, 0xe6, 0x6b, 0x77, 0x9c, 0xbe, 0xbd, 0x5e, 0x84, 0xee, 0xcd, 0x22, 0x74,
	0x7f, 0x2f, 0x42, 0xf7, 0xfb, 0x32, 0x74, 0x6e, 0x96, 0xa1, 0xf3, 0x6b, 0x19, 0x3a, 0x9f, 0x77,
	0xb9, 0xd0, 0x5f, 0xaa, 0x84, 0xa4, 0x38, 0xa3, 0xf6, 0xe9, 0x6b, 0x5e, 0xb2, 0x0c, 0xe8, 0x65,
	0xf3, 0x0f, 0xe8, 0x5a, 0x82, 0x4a, 0x06, 0xcd, 0x00, 0x5f, 0xff, 0x0b, 0x00, 0x00, 0xff, 0xff,
	0x99, 0xfd, 0xbe, 0xab, 0x1e, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// CreateValidator defines a method for creating a new validator.
	CreateValidator(ctx context.Context, in *MsgCreateValidator, opts ...grpc.CallOption) (*MsgCreateValidatorResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateValidator(ctx context.Context, in *MsgCreateValidator, opts ...grpc.CallOption) (*MsgCreateValidatorResponse, error) {
	out := new(MsgCreateValidatorResponse)
	err := c.cc.Invoke(ctx, "/confio.poe.v1beta1.Msg/CreateValidator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// CreateValidator defines a method for creating a new validator.
	CreateValidator(context.Context, *MsgCreateValidator) (*MsgCreateValidatorResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateValidator(ctx context.Context, req *MsgCreateValidator) (*MsgCreateValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateValidator not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CreateValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateValidator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateValidator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/confio.poe.v1beta1.Msg/CreateValidator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateValidator(ctx, req.(*MsgCreateValidator))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "confio.poe.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateValidator",
			Handler:    _Msg_CreateValidator_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "confio/poe/v1beta1/tx.proto",
}

func (m *MsgCreateValidator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateValidator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateValidator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Value.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	if m.Pubkey != nil {
		{
			size, err := m.Pubkey.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.DelegatorAddress) > 0 {
		i -= len(m.DelegatorAddress)
		copy(dAtA[i:], m.DelegatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.DelegatorAddress)))
		i--
		dAtA[i] = 0x22
	}
	{
		size, err := m.Description.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgCreateValidatorResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateValidatorResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateValidatorResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgCreateValidator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Description.Size()
	n += 1 + l + sovTx(uint64(l))
	l = len(m.DelegatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Pubkey != nil {
		l = m.Pubkey.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Value.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgCreateValidatorResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCreateValidator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateValidator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateValidator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Description.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DelegatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pubkey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pubkey == nil {
				m.Pubkey = &types1.Any{}
			}
			if err := m.Pubkey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Value.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCreateValidatorResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateValidatorResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateValidatorResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
