package godless

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

type replicator struct {
	topics   []RemoteStoreAddress
	interval time.Duration
	api      APIService
	store    RemoteStore
	errch    chan<- error
	stopch   <-chan interface{}
}

func (p2p replicator) publishAllTopics() {
	if len(p2p.topics) == 0 {
		loginfo("No topics, not publishing.")
		return
	}

	ticker := time.NewTicker(p2p.interval)
	defer ticker.Stop()
LOOP:
	for {
		select {
		case <-p2p.stopch:
			break LOOP
		case <-ticker.C:
			p2p.publishIndex()
		}
	}
}

func (p2p replicator) publishIndex() {
	loginfo("Publishing index...")

	addr, reflectErr := p2p.sendReflectRequest()

	if reflectErr != nil {
		logerr("Failed to reflect for index address: %v", reflectErr)
		return
	}

	// TODO ReflectResponse.Path should be RemoteStoreAddress.
	head := IPFSPath(addr.ReflectResponse.Path)

	publishErr := p2p.store.PublishAddr(head, p2p.topics)

	if publishErr != nil {
		logerr("Failed to publish index: %v", publishErr)
	}
}

func (p2p replicator) sendReflectRequest() (APIResponse, error) {
	failResp := RESPONSE_FAIL
	failResp.Type = API_REFLECT

	request := APIReflectRequest{Command: REFLECT_HEAD_PATH}
	respch, err := p2p.api.Reflect(request)

	if err != nil {
		return failResp, errors.Wrap(err, "Reflection failed (Early API failure) for: %v %v")
	}

	resp := <-respch
	loginfo("Reflection API Response message: %v", resp.Msg)

	if resp.Err != nil {
		return resp, resp.Err
	}

	return resp, nil
}

func (p2p replicator) subscribeAllTopics() {
	if len(p2p.topics) == 0 {
		loginfo("No topics, not subscribing.")
		return
	}

	wg := &sync.WaitGroup{}
	for _, t := range p2p.topics {
		topic := t
		wg.Add(1)
		go func() {
			p2p.subscribeTopic(topic)
			wg.Done()
		}()
	}

	wg.Wait()
}

func (p2p replicator) subscribeTopic(topic RemoteStoreAddress) {
	headch, errch := p2p.store.SubscribeAddrStream(topic)

	go func() {
	LOOP:
		for {
			select {
			case head, present := <-headch:
				if !present {
					break LOOP
				}

				p2p.sendReplicateRequest(head)
			case err, present := <-errch:
				if !present {
					break LOOP
				}
				loginfo("Subscription error: %v", err)
				break LOOP
			case <-p2p.stopch:
				break LOOP
			}
		}
	}()
}

func (p2p replicator) sendReplicateRequest(head RemoteStoreAddress) {
	loginfo("Replicating from: %v", head)

	respch, err := p2p.api.Replicate(head)

	if err != nil {
		logerr("Replication failed (Early API failure) for '%v': %v", head, err)
		return
	}

	resp := <-respch
	loginfo("Replication API Response message: %v", resp.Msg)
	if resp.Err != nil {
		logerr("Replication failed for '%v': %v", head, resp.Err)
	}
}

func Replicate(api APIService, store RemoteStore, interval time.Duration, topics []RemoteStoreAddress) (chan<- interface{}, <-chan error) {
	stopch := make(chan interface{}, 1)
	errch := make(chan error, len(topics))

	p2p := replicator{
		topics:   topics,
		interval: interval,
		api:      api,
		errch:    errch,
		stopch:   stopch,
		store:    store,
	}

	wg := &sync.WaitGroup{}
	wg.Add(__REPLICATION_PROCESS_COUNT)
	go func() {
		p2p.subscribeAllTopics()
		wg.Done()
	}()

	go func() {
		go p2p.publishAllTopics()
		wg.Done()
	}()

	return stopch, errch
}

const __REPLICATION_PROCESS_COUNT = 2
