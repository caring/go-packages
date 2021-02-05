// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pagination.proto

package pagination

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type PageInfo struct {
	HasNextPage          bool     `protobuf:"varint,1,opt,name=has_next_page,json=hasNextPage,proto3" json:"has_next_page,omitempty"`
	StartCursor          string   `protobuf:"bytes,2,opt,name=start_cursor,json=startCursor,proto3" json:"start_cursor,omitempty"`
	HasPreviousPage      bool     `protobuf:"varint,3,opt,name=has_previous_page,json=hasPreviousPage,proto3" json:"has_previous_page,omitempty"`
	EndCursor            string   `protobuf:"bytes,4,opt,name=end_cursor,json=endCursor,proto3" json:"end_cursor,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PageInfo) Reset()         { *m = PageInfo{} }
func (m *PageInfo) String() string { return proto.CompactTextString(m) }
func (*PageInfo) ProtoMessage()    {}
func (*PageInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_567bfb3a87c868dd, []int{0}
}

func (m *PageInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PageInfo.Unmarshal(m, b)
}
func (m *PageInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PageInfo.Marshal(b, m, deterministic)
}
func (m *PageInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PageInfo.Merge(m, src)
}
func (m *PageInfo) XXX_Size() int {
	return xxx_messageInfo_PageInfo.Size(m)
}
func (m *PageInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PageInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PageInfo proto.InternalMessageInfo

func (m *PageInfo) GetHasNextPage() bool {
	if m != nil {
		return m.HasNextPage
	}
	return false
}

func (m *PageInfo) GetStartCursor() string {
	if m != nil {
		return m.StartCursor
	}
	return ""
}

func (m *PageInfo) GetHasPreviousPage() bool {
	if m != nil {
		return m.HasPreviousPage
	}
	return false
}

func (m *PageInfo) GetEndCursor() string {
	if m != nil {
		return m.EndCursor
	}
	return ""
}

type PaginationRequest struct {
	First                int64    `protobuf:"varint,1,opt,name=first,proto3" json:"first,omitempty"`
	After                string   `protobuf:"bytes,2,opt,name=after,proto3" json:"after,omitempty"`
	Last                 int64    `protobuf:"varint,3,opt,name=last,proto3" json:"last,omitempty"`
	Before               string   `protobuf:"bytes,4,opt,name=before,proto3" json:"before,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PaginationRequest) Reset()         { *m = PaginationRequest{} }
func (m *PaginationRequest) String() string { return proto.CompactTextString(m) }
func (*PaginationRequest) ProtoMessage()    {}
func (*PaginationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_567bfb3a87c868dd, []int{1}
}

func (m *PaginationRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PaginationRequest.Unmarshal(m, b)
}
func (m *PaginationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PaginationRequest.Marshal(b, m, deterministic)
}
func (m *PaginationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PaginationRequest.Merge(m, src)
}
func (m *PaginationRequest) XXX_Size() int {
	return xxx_messageInfo_PaginationRequest.Size(m)
}
func (m *PaginationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PaginationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PaginationRequest proto.InternalMessageInfo

func (m *PaginationRequest) GetFirst() int64 {
	if m != nil {
		return m.First
	}
	return 0
}

func (m *PaginationRequest) GetAfter() string {
	if m != nil {
		return m.After
	}
	return ""
}

func (m *PaginationRequest) GetLast() int64 {
	if m != nil {
		return m.Last
	}
	return 0
}

func (m *PaginationRequest) GetBefore() string {
	if m != nil {
		return m.Before
	}
	return ""
}

func init() {
	proto.RegisterType((*PageInfo)(nil), "pagination.PageInfo")
	proto.RegisterType((*PaginationRequest)(nil), "pagination.PaginationRequest")
}

func init() { proto.RegisterFile("pagination.proto", fileDescriptor_567bfb3a87c868dd) }

var fileDescriptor_567bfb3a87c868dd = []byte{
	// 261 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0x89, 0xa9, 0xa5, 0x9d, 0x2a, 0xda, 0x45, 0xa4, 0x17, 0xa1, 0xe6, 0x54, 0x44, 0x93,
	0x83, 0x6f, 0xa0, 0x27, 0x2f, 0x12, 0x72, 0xf4, 0x12, 0x26, 0xe9, 0x64, 0xb3, 0x44, 0x77, 0xd7,
	0xdd, 0x89, 0xf4, 0x61, 0x7c, 0x58, 0xc9, 0x26, 0xb5, 0xb7, 0xfd, 0x3f, 0x96, 0x6f, 0x7e, 0x7e,
	0xb8, 0xb6, 0x28, 0x95, 0x46, 0x56, 0x46, 0xa7, 0xd6, 0x19, 0x36, 0x02, 0x4e, 0x24, 0xf9, 0x8d,
	0x60, 0x91, 0xa3, 0xa4, 0x37, 0xdd, 0x18, 0x91, 0xc0, 0x65, 0x8b, 0xbe, 0xd4, 0x74, 0xe0, 0xd2,
	0xa2, 0xa4, 0x4d, 0xb4, 0x8d, 0x76, 0x8b, 0x62, 0xd5, 0xa2, 0x7f, 0xa7, 0x03, 0x0f, 0xff, 0xc4,
	0x3d, 0x5c, 0x78, 0x46, 0xc7, 0x65, 0xdd, 0x3b, 0x6f, 0xdc, 0xe6, 0x6c, 0x1b, 0xed, 0x96, 0xc5,
	0x2a, 0xb0, 0xd7, 0x80, 0xc4, 0x03, 0xac, 0x07, 0x8d, 0x75, 0xf4, 0xa3, 0x4c, 0xef, 0x47, 0x55,
	0x1c, 0x54, 0x57, 0x2d, 0xfa, 0x7c, 0xe2, 0x41, 0x77, 0x07, 0x40, 0x7a, 0x7f, 0x94, 0xcd, 0x82,
	0x6c, 0x49, 0x7a, 0x3f, 0xaa, 0x92, 0x0e, 0xd6, 0xf9, 0x7f, 0xd9, 0x82, 0xbe, 0x7b, 0xf2, 0x2c,
	0x6e, 0xe0, 0xbc, 0x51, 0xce, 0x73, 0xa8, 0x17, 0x17, 0x63, 0x18, 0x28, 0x36, 0x4c, 0xc7, 0x46,
	0x63, 0x10, 0x02, 0x66, 0x9f, 0xe8, 0x39, 0x9c, 0x8f, 0x8b, 0xf0, 0x16, 0xb7, 0x30, 0xaf, 0xa8,
	0x31, 0x8e, 0xa6, 0x7b, 0x53, 0x7a, 0x49, 0x3f, 0x1e, 0xa5, 0xe2, 0xb6, 0xaf, 0xd2, 0xda, 0x7c,
	0x65, 0x35, 0x3a, 0xa5, 0x65, 0x26, 0xcd, 0x93, 0xc5, 0xba, 0x43, 0x49, 0x3e, 0xb3, 0x9d, 0xcc,
	0x4e, 0xdb, 0x55, 0xf3, 0x30, 0xe7, 0xf3, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x45, 0x0a, 0xa3,
	0x04, 0x62, 0x01, 0x00, 0x00,
}
