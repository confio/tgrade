// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: confio/twasm/v1beta1/proposal.proto

package types

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/CosmWasm/wasmd/x/wasm/types"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/regen-network/cosmos-proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

var (
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// PromoteToPrivilegedContractProposal gov proposal content type to add
// "privileges" to a contract
type PromoteToPrivilegedContractProposal struct {
	// Title is a short summary
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty" yaml:"title"`
	// Description is a human readable text
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty" yaml:"description"`
	// Contract is the address of the smart contract
	Contract string `protobuf:"bytes,3,opt,name=contract,proto3" json:"contract,omitempty" yaml:"contract"`
}

func (m *PromoteToPrivilegedContractProposal) Reset()      { *m = PromoteToPrivilegedContractProposal{} }
func (*PromoteToPrivilegedContractProposal) ProtoMessage() {}
func (*PromoteToPrivilegedContractProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_77ea8b6359ab7726, []int{0}
}

func (m *PromoteToPrivilegedContractProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *PromoteToPrivilegedContractProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PromoteToPrivilegedContractProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *PromoteToPrivilegedContractProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PromoteToPrivilegedContractProposal.Merge(m, src)
}

func (m *PromoteToPrivilegedContractProposal) XXX_Size() int {
	return m.Size()
}

func (m *PromoteToPrivilegedContractProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_PromoteToPrivilegedContractProposal.DiscardUnknown(m)
}

var xxx_messageInfo_PromoteToPrivilegedContractProposal proto.InternalMessageInfo

// PromoteToPrivilegedContractProposal gov proposal content type to remove
// "privileges" from a contract
type DemotePrivilegedContractProposal struct {
	// Title is a short summary
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty" yaml:"title"`
	// Description is a human readable text
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty" yaml:"description"`
	// Contract is the address of the smart contract
	Contract string `protobuf:"bytes,3,opt,name=contract,proto3" json:"contract,omitempty" yaml:"contract"`
}

func (m *DemotePrivilegedContractProposal) Reset()      { *m = DemotePrivilegedContractProposal{} }
func (*DemotePrivilegedContractProposal) ProtoMessage() {}
func (*DemotePrivilegedContractProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_77ea8b6359ab7726, []int{1}
}

func (m *DemotePrivilegedContractProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *DemotePrivilegedContractProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DemotePrivilegedContractProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *DemotePrivilegedContractProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DemotePrivilegedContractProposal.Merge(m, src)
}

func (m *DemotePrivilegedContractProposal) XXX_Size() int {
	return m.Size()
}

func (m *DemotePrivilegedContractProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_DemotePrivilegedContractProposal.DiscardUnknown(m)
}

var xxx_messageInfo_DemotePrivilegedContractProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*PromoteToPrivilegedContractProposal)(nil), "confio.twasm.v1beta1.PromoteToPrivilegedContractProposal")
	proto.RegisterType((*DemotePrivilegedContractProposal)(nil), "confio.twasm.v1beta1.DemotePrivilegedContractProposal")
}

func init() {
	proto.RegisterFile("confio/twasm/v1beta1/proposal.proto", fileDescriptor_77ea8b6359ab7726)
}

var fileDescriptor_77ea8b6359ab7726 = []byte{
	// 346 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x92, 0x4d, 0x4a, 0xc3, 0x40,
	0x14, 0x80, 0x33, 0x8a, 0xa2, 0x51, 0x50, 0x62, 0x91, 0x5a, 0x64, 0x5a, 0x52, 0x28, 0xae, 0x32,
	0x14, 0x37, 0xe2, 0xb2, 0xba, 0x74, 0x51, 0x8a, 0x2b, 0x37, 0x32, 0x49, 0xa7, 0xe3, 0x40, 0x92,
	0x17, 0x66, 0xa6, 0xd5, 0xde, 0xc2, 0x63, 0x78, 0x01, 0xc1, 0x23, 0x74, 0xd9, 0x65, 0x57, 0xc5,
	0xa6, 0x37, 0xe8, 0x09, 0xa4, 0x33, 0xd3, 0xd2, 0x2b, 0xb8, 0xcb, 0xcb, 0xf7, 0xbd, 0x9f, 0xe1,
	0x3d, 0xbf, 0x99, 0x40, 0x3e, 0x10, 0x40, 0xf4, 0x3b, 0x55, 0x19, 0x19, 0xb5, 0x63, 0xa6, 0x69,
	0x9b, 0x14, 0x12, 0x0a, 0x50, 0x34, 0x8d, 0x0a, 0x09, 0x1a, 0x82, 0x8a, 0x95, 0x22, 0x23, 0x45,
	0x4e, 0xaa, 0x55, 0x38, 0x70, 0x30, 0x02, 0x59, 0x7f, 0x59, 0xb7, 0x86, 0x13, 0x50, 0x19, 0x28,
	0x12, 0x53, 0xc5, 0xb6, 0xf5, 0x12, 0x10, 0xb9, 0xe3, 0xd7, 0x6b, 0x6e, 0x9a, 0xb9, 0x8e, 0x44,
	0x8f, 0x0b, 0xa6, 0x1c, 0xbd, 0xb2, 0xd9, 0xaf, 0xb6, 0xac, 0x0d, 0x36, 0x88, 0x03, 0xf0, 0x94,
	0x11, 0x13, 0xc5, 0xc3, 0x01, 0xa1, 0xf9, 0xd8, 0xa2, 0xf0, 0x07, 0xf9, 0xcd, 0xae, 0x84, 0x0c,
	0x34, 0x7b, 0x86, 0xae, 0x14, 0x23, 0x91, 0x32, 0xce, 0xfa, 0x0f, 0x90, 0x6b, 0x49, 0x13, 0xdd,
	0x75, 0xaf, 0x09, 0x5a, 0xfe, 0x81, 0x16, 0x3a, 0x65, 0x55, 0xd4, 0x40, 0x37, 0xc7, 0x9d, 0xf3,
	0xd5, 0xbc, 0x7e, 0x3a, 0xa6, 0x59, 0x7a, 0x1f, 0x9a, 0xdf, 0x61, 0xcf, 0xe2, 0xe0, 0xce, 0x3f,
	0xe9, 0x33, 0x95, 0x48, 0x51, 0x68, 0x01, 0x79, 0x75, 0xcf, 0xd8, 0x97, 0xab, 0x79, 0x3d, 0xb0,
	0xf6, 0x0e, 0x0c, 0x7b, 0xbb, 0x6a, 0x40, 0xfc, 0xa3, 0xc4, 0x75, 0xad, 0xee, 0x9b, 0xb4, 0x8b,
	0xd5, 0xbc, 0x7e, 0x66, 0xd3, 0x36, 0x24, 0xec, 0x6d, 0xa5, 0xf0, 0x1b, 0xf9, 0x8d, 0x47, 0xb6,
	0x9e, 0xfc, 0x5f, 0xcd, 0xdd, 0x79, 0x9a, 0x2c, 0xb0, 0x37, 0x5b, 0x60, 0xef, 0xab, 0xc4, 0x68,
	0x52, 0x62, 0x34, 0x2d, 0x31, 0xfa, 0x2d, 0x31, 0xfa, 0x5c, 0x62, 0x6f, 0xba, 0xc4, 0xde, 0x6c,
	0x89, 0xbd, 0x97, 0x16, 0x17, 0xfa, 0x6d, 0x18, 0x47, 0x09, 0x64, 0x64, 0x73, 0x68, 0x5c, 0xd2,
	0x3e, 0x23, 0x1f, 0xee, 0xe2, 0xcc, 0xf2, 0xe3, 0x43, 0xb3, 0xc7, 0xdb, 0xbf, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x6d, 0x82, 0xb6, 0xb6, 0x8e, 0x02, 0x00, 0x00,
}

func (this *PromoteToPrivilegedContractProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PromoteToPrivilegedContractProposal)
	if !ok {
		that2, ok := that.(PromoteToPrivilegedContractProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if this.Contract != that1.Contract {
		return false
	}
	return true
}

func (this *DemotePrivilegedContractProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DemotePrivilegedContractProposal)
	if !ok {
		that2, ok := that.(DemotePrivilegedContractProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if this.Contract != that1.Contract {
		return false
	}
	return true
}

func (m *PromoteToPrivilegedContractProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PromoteToPrivilegedContractProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PromoteToPrivilegedContractProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Contract) > 0 {
		i -= len(m.Contract)
		copy(dAtA[i:], m.Contract)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Contract)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DemotePrivilegedContractProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DemotePrivilegedContractProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DemotePrivilegedContractProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Contract) > 0 {
		i -= len(m.Contract)
		copy(dAtA[i:], m.Contract)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Contract)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func (m *PromoteToPrivilegedContractProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Contract)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func (m *DemotePrivilegedContractProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Contract)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}

func (m *PromoteToPrivilegedContractProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: PromoteToPrivilegedContractProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PromoteToPrivilegedContractProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Contract", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Contract = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
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

func (m *DemotePrivilegedContractProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: DemotePrivilegedContractProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DemotePrivilegedContractProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Contract", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Contract = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
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

func skipProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
				return 0, ErrInvalidLengthProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProposal = fmt.Errorf("proto: unexpected end of group")
)
