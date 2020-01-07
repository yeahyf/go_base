// Code generated by protoc-gen-go. DO NOT EDIT.
// source: errorresp.proto

package ept

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ErrorResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Info                 string   `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ErrorResponse) Reset()         { *m = ErrorResponse{} }
func (m *ErrorResponse) String() string { return proto.CompactTextString(m) }
func (*ErrorResponse) ProtoMessage()    {}
func (*ErrorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3749fe9bae333890, []int{0}
}

func (m *ErrorResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ErrorResponse.Unmarshal(m, b)
}
func (m *ErrorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ErrorResponse.Marshal(b, m, deterministic)
}
func (m *ErrorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrorResponse.Merge(m, src)
}
func (m *ErrorResponse) XXX_Size() int {
	return xxx_messageInfo_ErrorResponse.Size(m)
}
func (m *ErrorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ErrorResponse proto.InternalMessageInfo

func (m *ErrorResponse) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *ErrorResponse) GetInfo() string {
	if m != nil {
		return m.Info
	}
	return ""
}

func init() {
	proto.RegisterType((*ErrorResponse)(nil), "ept.ErrorResponse")
}

func init() { proto.RegisterFile("errorresp.proto", fileDescriptor_3749fe9bae333890) }

var fileDescriptor_3749fe9bae333890 = []byte{
	// 97 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x2d, 0x2a, 0xca,
	0x2f, 0x2a, 0x4a, 0x2d, 0x2e, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2d, 0x28,
	0x51, 0x32, 0xe7, 0xe2, 0x75, 0x05, 0x89, 0x07, 0xa5, 0x16, 0x17, 0xe4, 0xe7, 0x15, 0xa7, 0x0a,
	0x09, 0x71, 0xb1, 0x24, 0xe7, 0xa7, 0xa4, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xf0, 0x06, 0x81, 0xd9,
	0x20, 0xb1, 0xcc, 0xbc, 0xb4, 0x7c, 0x09, 0x26, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x30, 0x3b, 0x89,
	0x0d, 0x6c, 0x88, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xeb, 0xe1, 0xd7, 0x28, 0x57, 0x00, 0x00,
	0x00,
}
