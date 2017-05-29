package godless

type IndexStreamEntry struct {
	TableName TableName
	Links     []IPFSPath
}

func (entry IndexStreamEntry) Equals(other IndexStreamEntry) bool {
	if entry.TableName != other.TableName {
		return false
	}

	for i, myLink := range entry.Links {
		otherLink := other.Links[i]
		if myLink != otherLink {
			return false
		}
	}

	return true
}

func ReadIndexEntryMessage(message *IndexEntryMessage) IndexStreamEntry {
	entry := IndexStreamEntry{
		TableName: TableName(message.Table),
		Links:     make([]IPFSPath, len(message.Links)),
	}

	for i, l := range message.Links {
		entry.Links[i] = IPFSPath(l)
	}

	return entry
}

func MakeIndexEntryMessage(entry IndexStreamEntry) *IndexEntryMessage {
	message := &IndexEntryMessage{
		Table: string(entry.TableName),
		Links: make([]string, len(entry.Links)),
	}

	for i, l := range entry.Links {
		message.Links[i] = string(l)
	}

	return message
}

func MakeIndexStreamEntry(t TableName, addrs []RemoteStoreAddress) IndexStreamEntry {
	entry := IndexStreamEntry{
		TableName: t,
		Links:     make([]IPFSPath, len(addrs)),
	}

	for i, a := range addrs {
		entry.Links[i] = IPFSPath(a.Path())
	}

	return entry
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

	minSize := imin(len(a.Links), len(b.Links))
	for i := 0; i < minSize; i++ {
		al := a.Links[i]
		bl := a.Links[i]

		if al < bl {
			return true
		} else if al > bl {
			return false
		}
	}

	return len(a.Links) < len(b.Links)
}

func MakeIndexStream(index RemoteNamespaceIndex) []IndexStreamEntry {
	stream := make([]IndexStreamEntry, len(index.Index))

	i := 0
	for t, addrs := range index.Index {
		entry := MakeIndexStreamEntry(t, addrs)
		stream[i] = entry
		i++
	}

	return stream
}

func ReadIndexStream(stream []IndexStreamEntry) RemoteNamespaceIndex {
	index := EmptyRemoteNamespaceIndex()

	for _, entry := range stream {
		index = index.joinStreamEntry(entry)
	}

	return index
}

func MakeIndexStreamMessage(stream []IndexStreamEntry) *IndexMessage {
	message := &IndexMessage{Entries: make([]*IndexEntryMessage, len(stream))}

	for i, entry := range stream {
		message.Entries[i] = MakeIndexEntryMessage(entry)
	}

	return message
}

func ReadIndexStreamMessage(message *IndexMessage) []IndexStreamEntry {
	stream := make([]IndexStreamEntry, len(message.Entries))

	for i, emsg := range message.Entries {
		stream[i] = ReadIndexEntryMessage(emsg)
	}

	return stream
}
