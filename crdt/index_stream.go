package crdt

import (
	"sort"

	"github.com/johnny-morrice/godless/crypto"
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

	builder := &indexStreamBuilder{
		stream: make([]IndexStreamEntry, 0, count),
	}

	for t, addrs := range index.Index {
		builder.makeIndexStreamEntries(t, addrs)
	}

	builder.uniqueOrder()

	return builder.stream, builder.invalid
}

type indexStreamBuilder struct {
	stream  []IndexStreamEntry
	invalid []InvalidIndexEntry
}

func (builder *indexStreamBuilder) makeIndexStreamEntries(t TableName, addrs []Link) {
	for _, link := range addrs {
		path := link.Path()

		if len(link.Signatures()) == 0 {
			entry := IndexStreamEntry{
				TableName: t,
				Link:      path,
			}

			builder.appendEntry(entry)
			continue
		}

		for _, sig := range link.Signatures() {
			entry, err := MakeIndexStreamEntry(t, path, sig)
			if err == nil {
				builder.appendEntry(entry)
			} else {
				builder.appendInvalid(entry)
			}
		}
	}
}

func (builder *indexStreamBuilder) appendEntry(entry IndexStreamEntry) {
	builder.stream = append(builder.stream, entry)
}

func (builder *indexStreamBuilder) appendInvalid(entry IndexStreamEntry) {
	invalid := InvalidIndexEntry(entry)
	builder.invalid = append(builder.invalid, invalid)
}

func (builder *indexStreamBuilder) uniqueOrder() {
	sort.Sort(byIndexStreamOrder(builder.stream))
	builder.uniqSorted()
}

func (builder *indexStreamBuilder) uniqSorted() {
	if len(builder.stream) < 2 {
		return
	}

	uniqIndex := 0
	for i := 1; i < len(builder.stream); i++ {
		entry := builder.stream[i]
		last := builder.stream[uniqIndex]
		if entry != last {
			uniqIndex++
			builder.stream[uniqIndex] = entry
		}
	}

	builder.stream = builder.stream[:uniqIndex+1]
}

type InvalidIndexEntry IndexStreamEntry

func ReadIndexStream(stream []IndexStreamEntry) (Index, []InvalidIndexEntry) {
	index := EmptyIndex()

	invalid := make([]InvalidIndexEntry, 0, len(stream))

	for _, entry := range stream {
		err := index.addStreamEntry(entry)

		if err != nil {
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
		sigCount := len(link.Signatures())
		if sigCount > 0 {
			count = count + sigCount
		} else {
			count++
		}
	}

	return count
}
