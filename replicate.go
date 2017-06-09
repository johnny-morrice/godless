package godless

import (
	"sync"
	"time"
)

type replicator struct {
	topics   []RemoteStoreAddress
	interval time.Duration
	api      APIPeerService
	store    RemoteStore
	errch    chan<- error
	stopch   <-chan interface{}
}

func (p2p replicator) publishAllTopics() {

}

func (p2p replicator) subscribeAllTopics() {
	if len(p2p.topics) == 0 {
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
				if present {
					p2p.errch <- err
				}
				break LOOP
			case <-p2p.stopch:
				break LOOP
			}
		}
	}()
}

func (p2p replicator) sendReplicateRequest(head RemoteStoreAddress) {
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

func Replicate(api APIPeerService, store RemoteStore, interval time.Duration, topics []RemoteStoreAddress) (chan<- interface{}, <-chan error) {
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
