package crdt

import (
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/proto"
)

type IndexStreamEntry struct {
	TableName TableName
	Signature crypto.SignatureText
	Link      IPFSPath
}

func (entry IndexStreamEntry) Equals(other IndexStreamEntry) bool {
	ok := entry.TableName == other.TableName
	ok = ok && entry.Link == other.Link
	ok = ok && entry.Signature == other.Signature

	if !ok {
		return false
	}

	return true
}

func ReadIndexEntryMessage(message *proto.IndexEntryMessage) IndexStreamEntry {
	entry := IndexStreamEntry{
		TableName: TableName(message.Table),
		Link:      IPFSPath(message.Link),
		Signature: crypto.SignatureText(message.Signature),
	}

	return entry
}

func MakeIndexEntryMessage(entry IndexStreamEntry) *proto.IndexEntryMessage {
	message := &proto.IndexEntryMessage{
		Table:     string(entry.TableName),
		Link:      string(entry.Link),
		Signature: string(entry.Signature),
	}

	return message
}

// TODO does not support unsigned links.
func MakeIndexStreamEntries(t TableName, addrs []Link) ([]IndexStreamEntry, []InvalidIndexEntry) {
	count := countAddrEntries(addrs)

	entries := make([]IndexStreamEntry, 0, count)
	invalidEntries := []InvalidIndexEntry{}

	for _, link := range addrs {
		path := link.Path
		for _, sig := range link.Signatures {
			entry, err := MakeIndexStreamEntry(t, path, sig)
			if err == nil {
				entries = append(entries, entry)
			} else {
				invalidEntries = append(invalidEntries, InvalidIndexEntry(entry))
			}
		}
	}

	return entries, invalidEntries
}

func MakeIndexStreamEntry(t TableName, path IPFSPath, sig crypto.Signature) (IndexStreamEntry, error) {
	sigText, err := crypto.PrintSignature(sig)

	entry := IndexStreamEntry{
		TableName: t,
		Link:      path,
		Signature: sigText,
	}

	return entry, err
}

type byIndexStreamOrder []IndexStreamEntry

func (stream byIndexStreamOrder) Len() int {
	return len(stream)
}

func (stream byIndexStreamOrder) Swap(i, j int) {
	stream[i], stream[j] = stream[j], stream[i]
}

func (stream byIndexStreamOrder) Less(i, j int) bool {
	a := stream[i]
	b := stream[j]

	if a.TableName < b.TableName {
		return true
	} else if a.TableName > b.TableName {
		return false
	}

	if a.Link < b.Link {
		return true
	} else if a.Link > b.Link {
		return false
	}

	return a.Signature < b.Signature
}

func MakeIndexStream(index Index) ([]IndexStreamEntry, []InvalidIndexEntry) {
	count := 0

	for _, addrs := range index.Index {
		count = count + countAddrEntries(addrs)
	}

	stream := make([]IndexStreamEntry, count)
	invalidEntries := []InvalidIndexEntry{}

	i := 0
	for t, addrs := range index.Index {
		entries, invalid := MakeIndexStreamEntries(t, addrs)
		stream = append(stream, entries...)
		invalidEntries = append(invalidEntries, invalid...)
		i++
	}

	sort.Sort(byIndexStreamOrder(stream))

	return stream, invalidEntries
}

type InvalidIndexEntry IndexStreamEntry

func ReadIndexStream(stream []IndexStreamEntry) (Index, []InvalidIndexEntry) {
	index := EmptyIndex()

	invalid := make([]InvalidIndexEntry, 0, len(stream))

	for _, entry := range stream {
		var err error
		index, err = index.joinStreamEntry(entry)

		if err != nil {
			log.Error("Invalid stream entry")
			invalid = append(invalid, InvalidIndexEntry(entry))
		}
	}

	return index, invalid
}

func MakeIndexStreamMessage(stream []IndexStreamEntry) *proto.IndexMessage {
	message := &proto.IndexMessage{Entries: make([]*proto.IndexEntryMessage, len(stream))}

	for i, entry := range stream {
		message.Entries[i] = MakeIndexEntryMessage(entry)
	}

	return message
}

func ReadIndexStreamMessage(message *proto.IndexMessage) []IndexStreamEntry {
	stream := make([]IndexStreamEntry, len(message.Entries))

	for i, emsg := range message.Entries {
		stream[i] = ReadIndexEntryMessage(emsg)
	}

	return stream
}

func countAddrEntries(addrs []Link) int {
	count := 0

	for _, link := range addrs {
		sigCount := len(link.Signatures)
		if sigCount > 0 {
			count = count + sigCount
		} else {
			count++
		}
	}

	return count
}
