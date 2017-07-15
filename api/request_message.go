package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/proto"
	"github.com/johnny-morrice/godless/query"
)

func MakeRequestMessage(request Request) *proto.APIRequestMessage {
	message := &proto.APIRequestMessage{}
	message.Type = uint32(request.Type)

	message.Reflection = uint32(request.Reflection)

	message.Replicate = &proto.ReplicateMessage{}
	message.Replicate.Links = make([]*proto.LinkMessage, 0, len(request.Replicate))
	for _, link := range request.Replicate {
		lmsg, err := crdt.MakeLinkMessage(link)

		if err != nil {
			log.Error("Invalid Link: %s", err.Error())
			continue
		}

		message.Replicate.Links = append(message.Replicate.Links, lmsg)
	}

	if request.Query != nil {
		message.Query = query.MakeQueryMessage(request.Query)
	}

	return message
}

func ReadRequestMessage(message *proto.APIRequestMessage) Request {
	request := Request{}
	request.Type = MessageType(message.Type)
	request.Reflection = ReflectionType(message.Reflection)

	request.Replicate = make([]crdt.Link, 0, len(message.Replicate.Links))
	for _, lmsg := range message.Replicate.Links {
		link, err := crdt.ReadLinkMessage(lmsg)

		if err != nil {
			log.Error("Invalid LinkMessage: %s", err.Error())
			continue
		}

		request.Replicate = append(request.Replicate, link)
	}

	if message.Query != nil {
		query, err := query.ReadQueryMessage(message.Query)

		if err == nil {
			request.Query = query
		} else {
			log.Error("Invalid Query: %s", err.Error())

		}
	}

	return request
}
