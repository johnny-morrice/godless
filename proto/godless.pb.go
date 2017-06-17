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
	Table  string          `protobuf:"bytes,1,opt,name=table" json:"table,omitempty"`
	Row    string          `protobuf:"bytes,2,opt,name=row" json:"row,omitempty"`
	Entry  string          `protobuf:"bytes,3,opt,name=entry" json:"entry,omitempty"`
	Points []*PointMessage `protobuf:"bytes,4,rep,name=points" json:"points,omitempty"`
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

func (m *NamespaceEntryMessage) GetPoints() []*PointMessage {
	if m != nil {
		return m.Points
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
	// 657 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x4d, 0x6f, 0xd3, 0x4c,
	0x10, 0xd6, 0x36, 0x76, 0x5a, 0x4f, 0x5b, 0xbd, 0xed, 0xb6, 0x7d, 0xeb, 0x56, 0x15, 0x44, 0x3e,
	0x15, 0x55, 0x8a, 0x44, 0x10, 0x48, 0x70, 0xa2, 0xe5, 0x43, 0x6a, 0x25, 0x50, 0x30, 0x07, 0x24,
	0x38, 0x39, 0xc9, 0x92, 0x9a, 0x3a, 0x5e, 0xb3, 0xbb, 0x51, 0x9a, 0x2b, 0x7f, 0x03, 0xae, 0xfc,
	0x39, 0x7e, 0x05, 0xda, 0xf1, 0xae, 0x3f, 0x12, 0x5b, 0xe2, 0xe4, 0x19, 0xcf, 0xb3, 0xcf, 0x3c,
	0x3b, 0x33, 0x3b, 0xb0, 0x3b, 0xe5, 0x93, 0x84, 0x49, 0xd9, 0xcf, 0x04, 0x57, 0x9c, 0xba, 0xf8,
	0x09, 0x6e, 0x60, 0xef, 0x7d, 0x34, 0x63, 0x32, 0x8b, 0xc6, 0xec, 0x1d, 0x93, 0x32, 0x9a, 0x32,
	0xfa, 0x0c, 0x36, 0x59, 0xaa, 0x44, 0xcc, 0xa4, 0x4f, 0x7a, 0x9d, 0xf3, 0xed, 0xc1, 0x59, 0x7e,
	0xa6, 0x5f, 0x20, 0xdf, 0xa4, 0x4a, 0x2c, 0x0d, 0x3c, 0xb4, 0xe0, 0xe0, 0x07, 0x81, 0xa3, 0x46,
	0x08, 0x3d, 0x04, 0x57, 0x45, 0xa3, 0x84, 0xf9, 0xa4, 0x47, 0xce, 0xbd, 0x30, 0x77, 0xe8, 0x1e,
	0x74, 0x04, 0x5f, 0xf8, 0x1b, 0xf8, 0x4f, 0x9b, 0x1a, 0xa7, 0xc9, 0x96, 0x7e, 0x27, 0xc7, 0xa1,
	0x43, 0x2f, 0xa0, 0x9b, 0xf1, 0x38, 0x55, 0xd2, 0x77, 0x50, 0xce, 0x81, 0x91, 0x33, 0xd4, 0x3f,
	0xad, 0x0a, 0x03, 0x09, 0x5e, 0xc2, 0x4e, 0xf5, 0x3f, 0xa5, 0xe0, 0x28, 0x76, 0xaf, 0x4c, 0x66,
	0xb4, 0xe9, 0x19, 0x78, 0x32, 0x9e, 0xa6, 0x91, 0x9a, 0x0b, 0x66, 0xd2, 0x97, 0x3f, 0x82, 0x2b,
	0xd8, 0xb9, 0x4e, 0x27, 0xec, 0xde, 0x32, 0x0c, 0x56, 0xcb, 0xe1, 0x9b, 0xfc, 0x88, 0x6a, 0x2e,
	0xc5, 0x17, 0xd8, 0x5f, 0x8b, 0xb6, 0x54, 0x81, 0x82, 0x93, 0xc4, 0xe9, 0x9d, 0xd1, 0x81, 0x76,
	0x5d, 0x60, 0x67, 0x55, 0xe0, 0x1f, 0x02, 0xf4, 0x72, 0x78, 0x1d, 0x32, 0x99, 0xf1, 0x54, 0x16,
	0x6d, 0xf3, 0x61, 0x73, 0x96, 0x9b, 0x26, 0x81, 0x75, 0xb1, 0xac, 0x42, 0x70, 0x61, 0x72, 0xe4,
	0x0e, 0x56, 0x66, 0x99, 0xe5, 0xfc, 0xbb, 0x21, 0xda, 0xf4, 0x35, 0xec, 0x7e, 0x9f, 0x33, 0xb1,
	0xb4, 0xdc, 0xbe, 0xd3, 0x23, 0xe7, 0xdb, 0x83, 0x07, 0xe6, 0xc6, 0x97, 0xc3, 0xeb, 0x0f, 0xd5,
	0xb0, 0xbd, 0x77, 0xfd, 0x10, 0xbd, 0x81, 0xff, 0x04, 0xfb, 0x9a, 0xb0, 0xb1, 0x2a, 0x78, 0x5c,
	0xe4, 0xe9, 0x95, 0x3c, 0x61, 0x1d, 0x60, 0x99, 0x56, 0x0f, 0x06, 0x43, 0x38, 0x6e, 0xc9, 0x4a,
	0x9f, 0x82, 0x97, 0xda, 0x71, 0xc3, 0x2b, 0x6f, 0x0f, 0x8e, 0x57, 0x27, 0xd5, 0xf2, 0x96, 0xc8,
	0xe0, 0x37, 0x81, 0x93, 0x56, 0x01, 0x45, 0x55, 0x48, 0xa5, 0x2a, 0x14, 0x9c, 0x2c, 0x52, 0xb7,
	0xb6, 0x45, 0xda, 0xa6, 0x8f, 0xc0, 0x8d, 0x75, 0x87, 0xb1, 0x7c, 0xe5, 0x4c, 0x56, 0x27, 0x27,
	0xcc, 0x11, 0x75, 0x9d, 0xce, 0x3f, 0xeb, 0xfc, 0x45, 0x60, 0x07, 0xef, 0x6d, 0xa5, 0xfd, 0x0f,
	0x5d, 0x9e, 0xbd, 0xe2, 0x13, 0x2b, 0xce, 0x78, 0xe5, 0x5c, 0x6d, 0x54, 0xe7, 0xea, 0x02, 0x9c,
	0x6f, 0x3c, 0x4e, 0x8d, 0x3e, 0x9b, 0x10, 0x09, 0x6f, 0x78, 0x9c, 0xda, 0x84, 0x08, 0xa2, 0x8f,
	0xa1, 0x2b, 0x99, 0x2e, 0x87, 0xd1, 0x77, 0x52, 0x85, 0x7f, 0xc4, 0x48, 0xf1, 0xd0, 0x72, 0x60,
	0x70, 0x05, 0x7b, 0xab, 0x64, 0xb4, 0x0f, 0x8e, 0xe0, 0x0b, 0xfb, 0x4e, 0x4e, 0xab, 0x24, 0x21,
	0x5f, 0xd4, 0xd2, 0x6a, 0x5c, 0x30, 0x82, 0x83, 0x86, 0xa0, 0x5d, 0x0c, 0xa4, 0x5c, 0x0c, 0xcf,
	0xcb, 0x37, 0xb8, 0x81, 0xdc, 0x0f, 0x1b, 0xb8, 0x9b, 0x9f, 0xe2, 0x5b, 0xf0, 0xdb, 0x40, 0xe5,
	0xbe, 0x21, 0xd5, 0x7d, 0x73, 0x08, 0x2e, 0x2e, 0x13, 0x5b, 0x4f, 0x74, 0x82, 0xcf, 0x40, 0xd7,
	0xab, 0xa1, 0xb1, 0x49, 0x3c, 0x8b, 0x95, 0x69, 0x49, 0xee, 0xd0, 0x3e, 0xb8, 0x8b, 0x5b, 0x66,
	0x96, 0x4b, 0xb9, 0x30, 0xf0, 0xfc, 0x27, 0x1d, 0x28, 0x26, 0x04, 0x61, 0xc1, 0x4f, 0x02, 0xfb,
	0x6b, 0xc1, 0xd6, 0x7e, 0xbf, 0x00, 0x2f, 0x13, 0x6c, 0x12, 0x8f, 0x23, 0x65, 0x33, 0x9c, 0x55,
	0x33, 0x0c, 0x6d, 0xb0, 0x18, 0xaa, 0x02, 0xae, 0x97, 0xd9, 0x38, 0x89, 0xe6, 0x92, 0x49, 0xbf,
	0x53, 0x5b, 0x66, 0xeb, 0xda, 0x2c, 0x30, 0x58, 0xc0, 0x51, 0x23, 0x6f, 0xab, 0x40, 0x0a, 0xce,
	0x1d, 0x5b, 0xe6, 0xad, 0xf2, 0x42, 0xb4, 0xe9, 0x29, 0x6c, 0x25, 0xb1, 0x62, 0x22, 0x4a, 0xf2,
	0xcc, 0x5e, 0x58, 0xf8, 0x9a, 0x67, 0x2e, 0x99, 0x6e, 0xb9, 0x9e, 0xbe, 0xad, 0xd0, 0x78, 0xa3,
	0x2e, 0x4a, 0x7b, 0xf2, 0x37, 0x00, 0x00, 0xff, 0xff, 0xc4, 0x03, 0x2a, 0x07, 0xbb, 0x06, 0x00,
	0x00,
}
