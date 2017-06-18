// Code generated by protoc-gen-go. DO NOT EDIT.
// source: godless.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	godless.proto

It has these top-level messages:
	NamespaceMessage
	NamespaceEntryMessage
	PointMessage
	IndexMessage
	IndexEntryMessage
	APIResponseMessage
	APIQueryResponseMessage
	APIReflectResponseMessage
	QueryMessage
	QueryJoinMessage
	QueryRowJoinMessage
	QueryRowJoinEntryMessage
	QuerySelectMessage
	QueryWhereMessage
	QueryPredicateMessage
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type NamespaceMessage struct {
	Entries []*NamespaceEntryMessage `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty"`
}

func (m *NamespaceMessage) Reset()                    { *m = NamespaceMessage{} }
func (m *NamespaceMessage) String() string            { return proto1.CompactTextString(m) }
func (*NamespaceMessage) ProtoMessage()               {}
func (*NamespaceMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *NamespaceMessage) GetEntries() []*NamespaceEntryMessage {
	if m != nil {
		return m.Entries
	}
	return nil
}

type NamespaceEntryMessage struct {
	Table string        `protobuf:"bytes,1,opt,name=table" json:"table,omitempty"`
	Row   string        `protobuf:"bytes,2,opt,name=row" json:"row,omitempty"`
	Entry string        `protobuf:"bytes,3,opt,name=entry" json:"entry,omitempty"`
	Point *PointMessage `protobuf:"bytes,4,opt,name=point" json:"point,omitempty"`
}

func (m *NamespaceEntryMessage) Reset()                    { *m = NamespaceEntryMessage{} }
func (m *NamespaceEntryMessage) String() string            { return proto1.CompactTextString(m) }
func (*NamespaceEntryMessage) ProtoMessage()               {}
func (*NamespaceEntryMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *NamespaceEntryMessage) GetTable() string {
	if m != nil {
		return m.Table
	}
	return ""
}

func (m *NamespaceEntryMessage) GetRow() string {
	if m != nil {
		return m.Row
	}
	return ""
}

func (m *NamespaceEntryMessage) GetEntry() string {
	if m != nil {
		return m.Entry
	}
	return ""
}

func (m *NamespaceEntryMessage) GetPoint() *PointMessage {
	if m != nil {
		return m.Point
	}
	return nil
}

type PointMessage struct {
	Text      string `protobuf:"bytes,1,opt,name=text" json:"text,omitempty"`
	Signature string `protobuf:"bytes,2,opt,name=signature" json:"signature,omitempty"`
}

func (m *PointMessage) Reset()                    { *m = PointMessage{} }
func (m *PointMessage) String() string            { return proto1.CompactTextString(m) }
func (*PointMessage) ProtoMessage()               {}
func (*PointMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PointMessage) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *PointMessage) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

type IndexMessage struct {
	Entries []*IndexEntryMessage `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty"`
}

func (m *IndexMessage) Reset()                    { *m = IndexMessage{} }
func (m *IndexMessage) String() string            { return proto1.CompactTextString(m) }
func (*IndexMessage) ProtoMessage()               {}
func (*IndexMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *IndexMessage) GetEntries() []*IndexEntryMessage {
	if m != nil {
		return m.Entries
	}
	return nil
}

type IndexEntryMessage struct {
	Table     string `protobuf:"bytes,1,opt,name=table" json:"table,omitempty"`
	Link      string `protobuf:"bytes,2,opt,name=link" json:"link,omitempty"`
	Signature string `protobuf:"bytes,3,opt,name=signature" json:"signature,omitempty"`
}

func (m *IndexEntryMessage) Reset()                    { *m = IndexEntryMessage{} }
func (m *IndexEntryMessage) String() string            { return proto1.CompactTextString(m) }
func (*IndexEntryMessage) ProtoMessage()               {}
func (*IndexEntryMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *IndexEntryMessage) GetTable() string {
	if m != nil {
		return m.Table
	}
	return ""
}

func (m *IndexEntryMessage) GetLink() string {
	if m != nil {
		return m.Link
	}
	return ""
}

func (m *IndexEntryMessage) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

type APIResponseMessage struct {
	Message         string                     `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
	Error           string                     `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
	Type            uint32                     `protobuf:"varint,3,opt,name=type" json:"type,omitempty"`
	QueryResponse   *APIQueryResponseMessage   `protobuf:"bytes,4,opt,name=queryResponse" json:"queryResponse,omitempty"`
	ReflectResponse *APIReflectResponseMessage `protobuf:"bytes,5,opt,name=reflectResponse" json:"reflectResponse,omitempty"`
}

func (m *APIResponseMessage) Reset()                    { *m = APIResponseMessage{} }
func (m *APIResponseMessage) String() string            { return proto1.CompactTextString(m) }
func (*APIResponseMessage) ProtoMessage()               {}
func (*APIResponseMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *APIResponseMessage) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *APIResponseMessage) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *APIResponseMessage) GetType() uint32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *APIResponseMessage) GetQueryResponse() *APIQueryResponseMessage {
	if m != nil {
		return m.QueryResponse
	}
	return nil
}

func (m *APIResponseMessage) GetReflectResponse() *APIReflectResponseMessage {
	if m != nil {
		return m.ReflectResponse
	}
	return nil
}

type APIQueryResponseMessage struct {
	Namespace *NamespaceMessage `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
}

func (m *APIQueryResponseMessage) Reset()                    { *m = APIQueryResponseMessage{} }
func (m *APIQueryResponseMessage) String() string            { return proto1.CompactTextString(m) }
func (*APIQueryResponseMessage) ProtoMessage()               {}
func (*APIQueryResponseMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *APIQueryResponseMessage) GetNamespace() *NamespaceMessage {
	if m != nil {
		return m.Namespace
	}
	return nil
}

type APIReflectResponseMessage struct {
	Type      uint32            `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Path      string            `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	Index     *IndexMessage     `protobuf:"bytes,3,opt,name=index" json:"index,omitempty"`
	Namespace *NamespaceMessage `protobuf:"bytes,4,opt,name=namespace" json:"namespace,omitempty"`
}

func (m *APIReflectResponseMessage) Reset()                    { *m = APIReflectResponseMessage{} }
func (m *APIReflectResponseMessage) String() string            { return proto1.CompactTextString(m) }
func (*APIReflectResponseMessage) ProtoMessage()               {}
func (*APIReflectResponseMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *APIReflectResponseMessage) GetType() uint32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *APIReflectResponseMessage) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *APIReflectResponseMessage) GetIndex() *IndexMessage {
	if m != nil {
		return m.Index
	}
	return nil
}

func (m *APIReflectResponseMessage) GetNamespace() *NamespaceMessage {
	if m != nil {
		return m.Namespace
	}
	return nil
}

type QueryMessage struct {
	OpCode uint32              `protobuf:"varint,1,opt,name=opCode" json:"opCode,omitempty"`
	Table  string              `protobuf:"bytes,2,opt,name=table" json:"table,omitempty"`
	Join   *QueryJoinMessage   `protobuf:"bytes,3,opt,name=join" json:"join,omitempty"`
	Select *QuerySelectMessage `protobuf:"bytes,4,opt,name=select" json:"select,omitempty"`
}

func (m *QueryMessage) Reset()                    { *m = QueryMessage{} }
func (m *QueryMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryMessage) ProtoMessage()               {}
func (*QueryMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *QueryMessage) GetOpCode() uint32 {
	if m != nil {
		return m.OpCode
	}
	return 0
}

func (m *QueryMessage) GetTable() string {
	if m != nil {
		return m.Table
	}
	return ""
}

func (m *QueryMessage) GetJoin() *QueryJoinMessage {
	if m != nil {
		return m.Join
	}
	return nil
}

func (m *QueryMessage) GetSelect() *QuerySelectMessage {
	if m != nil {
		return m.Select
	}
	return nil
}

type QueryJoinMessage struct {
	Rows []*QueryRowJoinMessage `protobuf:"bytes,1,rep,name=rows" json:"rows,omitempty"`
}

func (m *QueryJoinMessage) Reset()                    { *m = QueryJoinMessage{} }
func (m *QueryJoinMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryJoinMessage) ProtoMessage()               {}
func (*QueryJoinMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *QueryJoinMessage) GetRows() []*QueryRowJoinMessage {
	if m != nil {
		return m.Rows
	}
	return nil
}

type QueryRowJoinMessage struct {
	Row     string                      `protobuf:"bytes,1,opt,name=row" json:"row,omitempty"`
	Entries []*QueryRowJoinEntryMessage `protobuf:"bytes,2,rep,name=entries" json:"entries,omitempty"`
}

func (m *QueryRowJoinMessage) Reset()                    { *m = QueryRowJoinMessage{} }
func (m *QueryRowJoinMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryRowJoinMessage) ProtoMessage()               {}
func (*QueryRowJoinMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *QueryRowJoinMessage) GetRow() string {
	if m != nil {
		return m.Row
	}
	return ""
}

func (m *QueryRowJoinMessage) GetEntries() []*QueryRowJoinEntryMessage {
	if m != nil {
		return m.Entries
	}
	return nil
}

type QueryRowJoinEntryMessage struct {
	Entry string `protobuf:"bytes,1,opt,name=entry" json:"entry,omitempty"`
	Point string `protobuf:"bytes,2,opt,name=point" json:"point,omitempty"`
}

func (m *QueryRowJoinEntryMessage) Reset()                    { *m = QueryRowJoinEntryMessage{} }
func (m *QueryRowJoinEntryMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryRowJoinEntryMessage) ProtoMessage()               {}
func (*QueryRowJoinEntryMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *QueryRowJoinEntryMessage) GetEntry() string {
	if m != nil {
		return m.Entry
	}
	return ""
}

func (m *QueryRowJoinEntryMessage) GetPoint() string {
	if m != nil {
		return m.Point
	}
	return ""
}

type QuerySelectMessage struct {
	Limit uint32             `protobuf:"varint,1,opt,name=limit" json:"limit,omitempty"`
	Where *QueryWhereMessage `protobuf:"bytes,2,opt,name=where" json:"where,omitempty"`
}

func (m *QuerySelectMessage) Reset()                    { *m = QuerySelectMessage{} }
func (m *QuerySelectMessage) String() string            { return proto1.CompactTextString(m) }
func (*QuerySelectMessage) ProtoMessage()               {}
func (*QuerySelectMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *QuerySelectMessage) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *QuerySelectMessage) GetWhere() *QueryWhereMessage {
	if m != nil {
		return m.Where
	}
	return nil
}

type QueryWhereMessage struct {
	OpCode    uint32                 `protobuf:"varint,1,opt,name=opCode" json:"opCode,omitempty"`
	Predicate *QueryPredicateMessage `protobuf:"bytes,2,opt,name=predicate" json:"predicate,omitempty"`
	Clauses   []*QueryWhereMessage   `protobuf:"bytes,3,rep,name=clauses" json:"clauses,omitempty"`
}

func (m *QueryWhereMessage) Reset()                    { *m = QueryWhereMessage{} }
func (m *QueryWhereMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryWhereMessage) ProtoMessage()               {}
func (*QueryWhereMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *QueryWhereMessage) GetOpCode() uint32 {
	if m != nil {
		return m.OpCode
	}
	return 0
}

func (m *QueryWhereMessage) GetPredicate() *QueryPredicateMessage {
	if m != nil {
		return m.Predicate
	}
	return nil
}

func (m *QueryWhereMessage) GetClauses() []*QueryWhereMessage {
	if m != nil {
		return m.Clauses
	}
	return nil
}

type QueryPredicateMessage struct {
	OpCode   uint32   `protobuf:"varint,1,opt,name=opCode" json:"opCode,omitempty"`
	Keys     []string `protobuf:"bytes,2,rep,name=keys" json:"keys,omitempty"`
	Literals []string `protobuf:"bytes,3,rep,name=literals" json:"literals,omitempty"`
	Userow   bool     `protobuf:"varint,4,opt,name=userow" json:"userow,omitempty"`
}

func (m *QueryPredicateMessage) Reset()                    { *m = QueryPredicateMessage{} }
func (m *QueryPredicateMessage) String() string            { return proto1.CompactTextString(m) }
func (*QueryPredicateMessage) ProtoMessage()               {}
func (*QueryPredicateMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *QueryPredicateMessage) GetOpCode() uint32 {
	if m != nil {
		return m.OpCode
	}
	return 0
}

func (m *QueryPredicateMessage) GetKeys() []string {
	if m != nil {
		return m.Keys
	}
	return nil
}

func (m *QueryPredicateMessage) GetLiterals() []string {
	if m != nil {
		return m.Literals
	}
	return nil
}

func (m *QueryPredicateMessage) GetUserow() bool {
	if m != nil {
		return m.Userow
	}
	return false
}

func init() {
	proto1.RegisterType((*NamespaceMessage)(nil), "proto.NamespaceMessage")
	proto1.RegisterType((*NamespaceEntryMessage)(nil), "proto.NamespaceEntryMessage")
	proto1.RegisterType((*PointMessage)(nil), "proto.PointMessage")
	proto1.RegisterType((*IndexMessage)(nil), "proto.IndexMessage")
	proto1.RegisterType((*IndexEntryMessage)(nil), "proto.IndexEntryMessage")
	proto1.RegisterType((*APIResponseMessage)(nil), "proto.APIResponseMessage")
	proto1.RegisterType((*APIQueryResponseMessage)(nil), "proto.APIQueryResponseMessage")
	proto1.RegisterType((*APIReflectResponseMessage)(nil), "proto.APIReflectResponseMessage")
	proto1.RegisterType((*QueryMessage)(nil), "proto.QueryMessage")
	proto1.RegisterType((*QueryJoinMessage)(nil), "proto.QueryJoinMessage")
	proto1.RegisterType((*QueryRowJoinMessage)(nil), "proto.QueryRowJoinMessage")
	proto1.RegisterType((*QueryRowJoinEntryMessage)(nil), "proto.QueryRowJoinEntryMessage")
	proto1.RegisterType((*QuerySelectMessage)(nil), "proto.QuerySelectMessage")
	proto1.RegisterType((*QueryWhereMessage)(nil), "proto.QueryWhereMessage")
	proto1.RegisterType((*QueryPredicateMessage)(nil), "proto.QueryPredicateMessage")
}

func init() { proto1.RegisterFile("godless.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 650 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xd6, 0x36, 0x4e, 0x5b, 0x4f, 0x13, 0x91, 0x6e, 0x5b, 0xea, 0x56, 0x11, 0x44, 0x3e, 0x05,
	0x21, 0x45, 0x22, 0x08, 0x24, 0x38, 0xd1, 0xf2, 0x23, 0x25, 0x12, 0x28, 0x98, 0x03, 0x12, 0x9c,
	0x9c, 0x64, 0x49, 0x4d, 0x1d, 0xaf, 0xd9, 0xdd, 0x28, 0xcd, 0x8d, 0xf7, 0x80, 0x2b, 0x2f, 0xc7,
	0x53, 0xa0, 0xfd, 0xf3, 0x4f, 0x62, 0x4b, 0x9c, 0x3c, 0xb3, 0xf3, 0xed, 0x37, 0xdf, 0xce, 0x8c,
	0x07, 0xda, 0x0b, 0x3a, 0x8f, 0x09, 0xe7, 0x83, 0x94, 0x51, 0x41, 0x71, 0x53, 0x7d, 0xfc, 0x31,
	0x74, 0x3e, 0x84, 0x4b, 0xc2, 0xd3, 0x70, 0x46, 0xde, 0x13, 0xce, 0xc3, 0x05, 0xc1, 0xcf, 0xe1,
	0x80, 0x24, 0x82, 0x45, 0x84, 0x7b, 0xa8, 0xd7, 0xe8, 0x1f, 0x0d, 0xbb, 0xfa, 0xce, 0x20, 0x43,
	0xbe, 0x4d, 0x04, 0xdb, 0x18, 0x78, 0x60, 0xc1, 0xfe, 0x4f, 0x04, 0x67, 0x95, 0x10, 0x7c, 0x0a,
	0x4d, 0x11, 0x4e, 0x63, 0xe2, 0xa1, 0x1e, 0xea, 0xbb, 0x81, 0x76, 0x70, 0x07, 0x1a, 0x8c, 0xae,
	0xbd, 0x3d, 0x75, 0x26, 0x4d, 0x89, 0x93, 0x64, 0x1b, 0xaf, 0xa1, 0x71, 0xca, 0xc1, 0x8f, 0xa0,
	0x99, 0xd2, 0x28, 0x11, 0x9e, 0xd3, 0x43, 0xfd, 0xa3, 0xe1, 0x89, 0x51, 0x33, 0x91, 0x67, 0x56,
	0x84, 0x46, 0xf8, 0xaf, 0xa0, 0x55, 0x3c, 0xc6, 0x18, 0x1c, 0x41, 0xee, 0x84, 0xc9, 0xab, 0x6c,
	0xdc, 0x05, 0x97, 0x47, 0x8b, 0x24, 0x14, 0x2b, 0x46, 0x4c, 0xf2, 0xfc, 0xc0, 0xbf, 0x86, 0xd6,
	0x28, 0x99, 0x93, 0x3b, 0xcb, 0x30, 0xdc, 0x2e, 0x86, 0x67, 0xd2, 0x2b, 0x54, 0x75, 0x21, 0xbe,
	0xc2, 0xf1, 0x4e, 0xb4, 0xa6, 0x06, 0x18, 0x9c, 0x38, 0x4a, 0x6e, 0x8d, 0x0e, 0x65, 0x97, 0x05,
	0x36, 0xb6, 0x05, 0xfe, 0x45, 0x80, 0xaf, 0x26, 0xa3, 0x80, 0xf0, 0x94, 0x26, 0x3c, 0x6b, 0x9a,
	0x07, 0x07, 0x4b, 0x6d, 0x9a, 0x04, 0xd6, 0x55, 0x45, 0x65, 0x8c, 0x32, 0x93, 0x43, 0x3b, 0xaa,
	0x32, 0x9b, 0x54, 0xf3, 0xb7, 0x03, 0x65, 0xe3, 0x37, 0xd0, 0xfe, 0xb1, 0x22, 0x6c, 0x63, 0xb9,
	0x4d, 0xc1, 0x1f, 0x98, 0x17, 0x5f, 0x4d, 0x46, 0x1f, 0x8b, 0x61, 0xfb, 0xee, 0xf2, 0x25, 0x3c,
	0x86, 0x7b, 0x8c, 0x7c, 0x8b, 0xc9, 0x4c, 0x64, 0x3c, 0x4d, 0xc5, 0xd3, 0xcb, 0x79, 0x82, 0x32,
	0xc0, 0x32, 0x6d, 0x5f, 0xf4, 0x27, 0x70, 0x5e, 0x93, 0x15, 0x3f, 0x03, 0x37, 0xb1, 0xc3, 0xa6,
	0x9e, 0x7c, 0x34, 0x3c, 0xdf, 0x9e, 0x53, 0xcb, 0x9b, 0x23, 0xfd, 0x3f, 0x08, 0x2e, 0x6a, 0x05,
	0x64, 0x55, 0x41, 0x85, 0xaa, 0x60, 0x70, 0xd2, 0x50, 0xdc, 0xd8, 0x16, 0x49, 0x5b, 0x8e, 0x64,
	0x24, 0x3b, 0xac, 0xca, 0x97, 0x8f, 0x64, 0x71, 0x72, 0x02, 0x8d, 0x28, 0xeb, 0x74, 0xfe, 0x5b,
	0xe7, 0x6f, 0x04, 0x2d, 0xf5, 0x6e, 0x2b, 0xed, 0x3e, 0xec, 0xd3, 0xf4, 0x35, 0x9d, 0x5b, 0x71,
	0xc6, 0xcb, 0xe7, 0x6a, 0xaf, 0x38, 0x57, 0x8f, 0xc1, 0xf9, 0x4e, 0xa3, 0xc4, 0xe8, 0xb3, 0x09,
	0x15, 0xe1, 0x98, 0x46, 0x89, 0x4d, 0xa8, 0x40, 0xf8, 0x09, 0xec, 0x73, 0x22, 0xcb, 0x61, 0xf4,
	0x5d, 0x14, 0xe1, 0x9f, 0x54, 0xc4, 0x5e, 0x30, 0x40, 0xff, 0x1a, 0x3a, 0xdb, 0x64, 0x78, 0x00,
	0x0e, 0xa3, 0x6b, 0xfb, 0x9f, 0x5c, 0x16, 0x49, 0x02, 0xba, 0x2e, 0xa5, 0x95, 0x38, 0x7f, 0x0a,
	0x27, 0x15, 0x41, 0xbb, 0x16, 0x50, 0xbe, 0x16, 0x5e, 0xe4, 0xff, 0xe0, 0x9e, 0xe2, 0x7e, 0x58,
	0xc1, 0x5d, 0xfd, 0x2b, 0xbe, 0x03, 0xaf, 0x0e, 0x94, 0x6f, 0x1b, 0x54, 0xdc, 0x36, 0xa7, 0x76,
	0xdb, 0x98, 0x7a, 0xea, 0xc5, 0xf2, 0x05, 0xf0, 0x6e, 0x35, 0x24, 0x36, 0x8e, 0x96, 0x91, 0x30,
	0x2d, 0xd1, 0x0e, 0x1e, 0x40, 0x73, 0x7d, 0x43, 0xcc, 0x72, 0xc9, 0x17, 0x86, 0xba, 0xff, 0x59,
	0x06, 0xb2, 0x09, 0x51, 0x30, 0xff, 0x17, 0x82, 0xe3, 0x9d, 0x60, 0x6d, 0xbf, 0x5f, 0x82, 0x9b,
	0x32, 0x32, 0x8f, 0x66, 0xa1, 0xb0, 0x19, 0xba, 0xc5, 0x0c, 0x13, 0x1b, 0xcc, 0x86, 0x2a, 0x83,
	0xcb, 0x65, 0x36, 0x8b, 0xc3, 0x15, 0x27, 0xdc, 0x6b, 0x94, 0x96, 0xd9, 0xae, 0x36, 0x0b, 0xf4,
	0xd7, 0x70, 0x56, 0xc9, 0x5b, 0x2b, 0x10, 0x83, 0x73, 0x4b, 0x36, 0xba, 0x55, 0x6e, 0xa0, 0x6c,
	0x7c, 0x09, 0x87, 0x71, 0x24, 0x08, 0x0b, 0x63, 0x9d, 0xd9, 0x0d, 0x32, 0x5f, 0xf2, 0xac, 0x38,
	0x91, 0x2d, 0x97, 0xd3, 0x77, 0x18, 0x18, 0x6f, 0xba, 0xaf, 0xa4, 0x3d, 0xfd, 0x17, 0x00, 0x00,
	0xff, 0xff, 0xcc, 0xed, 0x56, 0x35, 0xb9, 0x06, 0x00, 0x00,
}
