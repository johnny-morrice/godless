package datapeer

import (
	"io"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/api"
)

type Union struct {
	Storage      api.ContentAddressableStorage
	Publisher    api.PubSubPublisher
	Subscriber   api.PubSubSubscriber
	Connecter    api.ConnectablePeer
	Disconnecter api.DisconnectablePeer
	Pinger       api.PingablePeer
}

func (peer Union) IsUp() bool {
	if peer.Pinger == nil {
		return true
	}

	return peer.Pinger.IsUp()
}

func (peer Union) Connect() error {
	if peer.Connecter == nil {
		return nil
	}

	return peer.Connecter.Connect()
}

func (peer Union) Disconnect() error {
	if peer.Disconnecter == nil {
		return nil
	}

	return peer.Disconnecter.Disconnect()
}

func (peer Union) Cat(path string) (io.ReadCloser, error) {
	if peer.Storage == nil {
		return nil, errors.New("no implementation added to datapeer.Union")
	}

	return peer.Storage.Cat(path)
}

func (peer Union) Add(r io.Reader) (string, error) {
	if peer.Storage == nil {
		return "", unionError
	}

	return peer.Storage.Add(r)
}

func (peer Union) PubSubPublish(topic, data string) error {
	if peer.Publisher == nil {
		return unionError
	}

	return peer.Publisher.PubSubPublish(topic, data)
}

func (peer Union) PubSubSubscribe(topic string) (api.PubSubSubscription, error) {
	if peer.Subscriber == nil {
		return nil, unionError
	}

	return peer.Subscriber.PubSubSubscribe(topic)
}

var unionError error = errors.New("no implementation added to datapeer.Union")
