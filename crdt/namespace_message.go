package crdt

import (
	"io"

	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/internal/debug"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

func ReadNamespaceEntryMessage(message *proto.NamespaceEntryMessage) NamespaceStreamEntry {
	entry := NamespaceStreamEntry{
		Table: TableName(message.Table),
		Row:   RowName(message.Row),
		Entry: EntryName(message.Entry),
		Point: ReadPointMessage(message.Point),
	}

	return entry
}

func ReadPointMessage(message *proto.PointMessage) StreamPoint {
	return StreamPoint{
		Text:      PointText(message.Text),
		Signature: crypto.SignatureText(message.Signature),
	}
}

func MakePointMessage(point StreamPoint) *proto.PointMessage {
	return &proto.PointMessage{
		Text:      string(point.Text),
		Signature: string(point.Signature),
	}
}

func MakeNamespaceEntryMessage(entry NamespaceStreamEntry) *proto.NamespaceEntryMessage {
	pb := &proto.NamespaceEntryMessage{
		Table: string(entry.Table),
		Row:   string(entry.Row),
		Entry: string(entry.Entry),
		Point: MakePointMessage(entry.Point),
	}

	return pb
}

func MakeNamespaceStreamMessage(stream []NamespaceStreamEntry) *proto.NamespaceMessage {
	message := &proto.NamespaceMessage{Entries: make([]*proto.NamespaceEntryMessage, len(stream))}

	for i, entry := range stream {
		message.Entries[i] = MakeNamespaceEntryMessage(entry)
	}

	debug.AssertLenEquals(stream, message.Entries)

	return message
}

func ReadNamespaceStreamMessage(message *proto.NamespaceMessage) []NamespaceStreamEntry {
	stream := make([]NamespaceStreamEntry, len(message.Entries))

	for i, emsg := range message.Entries {
		stream[i] = ReadNamespaceEntryMessage(emsg)
	}

	debug.AssertLenEquals(message.Entries, stream)

	return stream
}

func ReadNamespaceMessage(message *proto.NamespaceMessage) Namespace {
	stream := ReadNamespaceStreamMessage(message)
	return ReadNamespaceStream(stream)
}

func MakeNamespaceMessage(ns Namespace) *proto.NamespaceMessage {
	stream := MakeNamespaceStream(ns)
	return MakeNamespaceStreamMessage(stream)
}

func EncodeNamespace(ns Namespace, w io.Writer) error {
	const failMsg = "EncodeNamespace failed"

	message := MakeNamespaceMessage(ns)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeNamespace(r io.Reader) (Namespace, error) {
	const failMsg = "DecodeNamespace failed"
	message := &proto.NamespaceMessage{}
	err := util.Decode(message, r)

	if err != nil {
		return EmptyNamespace(), errors.Wrap(err, failMsg)
	}

	return ReadNamespaceMessage(message), nil
}
