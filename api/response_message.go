package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

func MakeAPIResponseMessage(resp Response) *proto.APIResponseMessage {
	message := &proto.APIResponseMessage{
		Message: resp.Msg,
		Type:    uint32(resp.Type),
		Path:    string(resp.Path),
	}

	if resp.Err != nil {
		message.Error = resp.Err.Error()
		return message
	}

	nsMsg, nsInvalid := crdt.MakeNamespaceMessage(resp.Namespace)
	message.Namespace = nsMsg
	logInvalidNamespace(nsInvalid)

	indexMsg, indexInvalid := crdt.MakeIndexMessage(resp.Index)
	message.Index = indexMsg

	logInvalidIndex(indexInvalid)

	return message
}

func ReadAPIResponseMessage(message *proto.APIResponseMessage) Response {
	resp := Response{
		Msg:  message.Message,
		Type: MessageType(message.Type),
		Path: crdt.IPFSPath(message.Path),
	}

	if message.Error != "" {
		resp.Err = errors.New(message.Error)
		return resp
	}

	if message.Namespace != nil {
		ns, nsInvalid := crdt.ReadNamespaceMessage(message.Namespace)
		resp.Namespace = ns
		logInvalidNamespace(nsInvalid)
	}

	if message.Index != nil {
		index, indexInvalid := crdt.ReadIndexMessage(message.Index)
		resp.Index = index
		logInvalidIndex(indexInvalid)
	}

	return resp
}

func logInvalidNamespace(invalid []crdt.InvalidNamespaceEntry) {
	if len(invalid) > 0 {
		log.Error("%d invalid Namespace entries", len(invalid))
	}
}

func logInvalidIndex(invalid []crdt.InvalidIndexEntry) {
	if len(invalid) > 0 {
		log.Error("%d invalid Index entries", len(invalid))
	}
}
