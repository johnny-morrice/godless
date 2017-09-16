package datapeer

import (
	"io"

	"github.com/johnny-morrice/godless/api"
)

type ProfilingDataPeer struct {
	ProfilingDataPeerOptions

	timerList *api.TimerList
}

type ProfilingDataPeerOptions struct {
	Peer api.DataPeer
	Prof api.Profiler
}

func MakeProfilingDataPeer(options ProfilingDataPeerOptions) api.DataPeer {
	if options.Peer == nil {
		panic("options.Peer was nil")
	}

	if options.Prof == nil {
		panic("options.Prof was nil")
	}

	return &ProfilingDataPeer{
		ProfilingDataPeerOptions: options,
		timerList:                &api.TimerList{},
	}
}

func (peer *ProfilingDataPeer) IsUp() bool {
	const profileName = __PROFILING_DATA_PEER_NAME + ".IsUp"

	var up bool
	peer.withTimer(profileName, func() {
		up = peer.Peer.IsUp()
	})

	return up
}

func (peer *ProfilingDataPeer) Connect() error {
	const profileName = __PROFILING_DATA_PEER_NAME + ".Connect"

	var err error
	peer.withTimer(profileName, func() {
		err = peer.Peer.Connect()
	})

	return err
}

func (peer *ProfilingDataPeer) Disconnect() error {
	const profileName = __PROFILING_DATA_PEER_NAME + ".Disconnect"

	defer peer.timerList.StopAllTimers()

	var err error
	peer.withTimer(profileName, func() {
		err = peer.Peer.Disconnect()
	})

	return err
}

func (peer *ProfilingDataPeer) Cat(hash string) (io.ReadCloser, error) {
	const profileName = __PROFILING_DATA_PEER_NAME + ".Cat"

	var reader io.ReadCloser
	var err error
	peer.withTimer(profileName, func() {
		reader, err = peer.Peer.Cat(hash)
	})

	return reader, err
}

func (peer *ProfilingDataPeer) Add(r io.Reader) (string, error) {
	const profileName = __PROFILING_DATA_PEER_NAME + ".Add"

	var path string
	var err error
	peer.withTimer(profileName, func() {
		path, err = peer.Peer.Add(r)
	})

	return path, err
}

func (peer *ProfilingDataPeer) PubSubPublish(topic, data string) error {
	const profileName = __PROFILING_DATA_PEER_NAME + ".PubSubPublish"

	var err error
	peer.withTimer(profileName, func() {
		err = peer.Peer.PubSubPublish(topic, data)
	})

	return err
}

func (peer *ProfilingDataPeer) PubSubSubscribe(topic string) (api.PubSubSubscription, error) {
	const profileName = __PROFILING_DATA_PEER_NAME + ".PubSubSubscribe"

	var plainSub api.PubSubSubscription
	var sub api.PubSubSubscription
	var err error
	peer.withTimer(profileName, func() {

		plainSub, err = peer.Peer.PubSubSubscribe(topic)
	})

	if err == nil {
		sub = profilingPubSubSubscription{
			peer: peer,
			sub:  plainSub,
		}
	}

	return sub, err
}

func (peer *ProfilingDataPeer) withTimer(name string, f func()) {
	timer := peer.Prof.NewTimer(name)
	peer.timerList.StartTimer(timer)
	defer peer.timerList.StopTimer(timer)
	f()
}

type profilingPubSubSubscription struct {
	peer *ProfilingDataPeer
	sub  api.PubSubSubscription
}

func (sub profilingPubSubSubscription) Next() (api.PubSubRecord, error) {
	const profileName = __PROFILING_PUB_SUB_SUBSCRIPTION + ".Next"

	var record api.PubSubRecord
	var err error
	sub.peer.withTimer(profileName, func() {
		record, err = sub.sub.Next()
	})

	return record, err
}

const __PROFILING_PUB_SUB_SUBSCRIPTION = "ProfilingPubSubSubscription"
const __PROFILING_DATA_PEER_NAME = "ProfilingDataPeer"
