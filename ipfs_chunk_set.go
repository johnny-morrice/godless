package godless

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"github.com/pkg/errors"
)

type IpfsChunkSet struct {
	loaded bool
	FilePeer *IpfsPeer
	FilePath IpfsPath
	ChunkSet *ChunkSet
	Children []*IpfsChunkSet
}

type IpfsChunkSetRecord struct {
	ChunkSet *ChunkSet
	Children []IpfsPath
}

func (ics *IpfsChunkSet) LoadChunks() (*ChunkSet, error) {
	out := &ChunkSet{}

	err := ics.LoadTraverse(func (child *IpfsChunkSet) {
		out = out.Join(child.ChunkSet)
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error in IpfsChunkSet LoadChunks")
	}

	return out, nil
}

func (ics *IpfsChunkSet) LoadTraverse(f func (*IpfsChunkSet)) error {
	stack := make([]*IpfsChunkSet, 1)
	stack[0] = ics

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.Load()

		if err != nil {
			return errors.Wrap(err, "Error in IpfsChunkSet LoadTraverse")
		}

		f(current)
		stack = append(stack, current.Children...)
	}

	return nil
}

// Load chunks over IPFS
func (ics *IpfsChunkSet) Load() error {
	if ics.FilePath == "" {
		panic("Tried to load IpfsChunkSet with empty FilePath")
	}

	if ics.loaded {
		log.Printf("WARN IpfsChunkSet already loaded from: '%v'", ics.FilePath)
		return nil
	}

	reader, err := ics.FilePeer.Shell.Cat(string(ics.FilePath))

	if err != nil {
		return errors.Wrap(err, "Error in IpfsChunkSet Cat")
	}

	defer reader.Close()

	dec := gob.NewDecoder(reader)
	part := &IpfsChunkSetRecord{}
	err = dec.Decode(part)

	// According to IPFS binding docs we must drain the reader.
	remainder, drainerr := ioutil.ReadAll(reader)

	if (drainerr != nil) {
		log.Printf("WARN error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		log.Printf("WARN remaining bits after gob: %v", remainder)
	}

	ics.ChunkSet = part.ChunkSet
	ics.Children = make([]*IpfsChunkSet, len(part.Children))

	for i, file := range part.Children {
		child := &IpfsChunkSet{}
		child.FilePath = file
		child.FilePeer = ics.FilePeer
		ics.Children[i] = child
	}

	if err != nil {
		return errors.Wrap(err, "Error decoding IpfsChunkSet Gob")
	}

	ics.loaded = true
	return nil
}

// Write chunks to IPFS
func (ics *IpfsChunkSet) Persist() error {
	if ics.loaded {
		log.Printf("WARN persisting loaded IpfsChunkset at: %v", ics.FilePath)
	}

	if ics.FilePath != "" {
		log.Printf("WARN IpfsChunkSet already persisted at: %v", ics.FilePath)
		return nil
	}

	part := &IpfsChunkSetRecord{}
	part.ChunkSet = ics.ChunkSet

	part.Children = make([]IpfsPath, len(ics.Children))
	for i, child := range ics.Children {
		part.Children[i] = child.FilePath
	}

	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(part)

	if err != nil {
		return errors.Wrap(err, "Error encoding IpfsChunkSet Gob")
	}

	addr, sherr := ics.FilePeer.Shell.Add(&buff)

	if sherr != nil {
		return errors.Wrap(err, "Error adding IpfsChunkSet to Ipfs")
	}

	ics.FilePath = IpfsPath(addr)

	return nil
}
