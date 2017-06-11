package crdt

import (
	"io"

	"github.com/johnny-morrice/godless/internal/debug"
	"github.com/johnny-morrice/godless/proto"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/pkg/errors"
)

func ReadNamespaceEntryMessage(message *proto.NamespaceEntryMessage) NamespaceStreamEntry {
	entry := NamespaceStreamEntry{
		Table:  TableName(message.Table),
		Row:    RowName(message.Row),
		Entry:  EntryName(message.Entry),
		Points: make([]Point, len(message.Points)),
	}

	for i, p := range message.Points {
		entry.Points[i] = Point(p)
	}

	return entry
}

func MakeNamespaceEntryMessage(entry NamespaceStreamEntry) *proto.NamespaceEntryMessage {
	pb := &proto.NamespaceEntryMessage{
		Table:  string(entry.Table),
		Row:    string(entry.Row),
		Entry:  string(entry.Entry),
		Points: make([]string, len(entry.Points)),
	}

	for i, p := range entry.Points {
		pb.Points[i] = string(p)
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
