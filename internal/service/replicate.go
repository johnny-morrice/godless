package service

import (
	"sync"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type replicator struct {
	topics   []crdt.IPFSPath
	interval time.Duration
	api      api.APIService
	store    api.RemoteStore
	errch    chan<- error
	stopch   <-chan interface{}
}

func (p2p replicator) publishAllTopics() {
	if len(p2p.topics) == 0 {
		log.Info("No topics, not publishing.")
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
	log.Info("Publishing index...")

	addr, reflectErr := p2p.sendReflectRequest()

	if reflectErr != nil {
		log.Error("Failed to reflect for index address: %v", reflectErr)
		return
	}

	// TODO ReflectResponse.Path should be crdt.IPFSPath.
	head := crdt.IPFSPath(addr.ReflectResponse.Path)

	publishErr := p2p.store.PublishAddr(head, p2p.topics)

	if publishErr != nil {
		log.Error("Failed to publish index: %v", publishErr)
	}
}

func (p2p replicator) sendReflectRequest() (api.APIResponse, error) {
	failResp := api.RESPONSE_FAIL
	failResp.Type = api.API_REFLECT

	respch, err := p2p.api.Reflect(api.REFLECT_HEAD_PATH)

	if err != nil {
		return failResp, errors.Wrap(err, "Reflection failed (Early API failure) for: %v %v")
	}

	resp := <-respch
	log.Info("Reflection API Response message: %v", resp.Msg)

	if resp.Err != nil {
		return resp, resp.Err
	}

	return resp, nil
}

func (p2p replicator) subscribeAllTopics() {
	if len(p2p.topics) == 0 {
		log.Info("No topics, not subscribing.")
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

func (p2p replicator) subscribeTopic(topic crdt.IPFSPath) {
	headch, errch := p2p.store.SubscribeAddrStream(topic)

	ticker := time.NewTicker(p2p.interval)
	defer ticker.Stop()
	go func() {
	LOOP:
		for {
			select {
			case head, present := <-headch:
				if !present {
					break LOOP
				}

				<-ticker.C
				p2p.sendReplicateRequest(head)
			case err, present := <-errch:
				if !present {
					break LOOP
				}
				log.Info("Subscription error: %v", err)
				break LOOP
			case <-p2p.stopch:
				break LOOP
			}
		}
	}()
}

func (p2p replicator) sendReplicateRequest(head crdt.IPFSPath) {
	log.Info("Replicating from: %v", head)

	respch, err := p2p.api.Replicate(head)

	if err != nil {
		log.Error("Replication failed (Early API failure) for '%v': %v", head, err)
		return
	}

	resp := <-respch
	log.Info("Replication API Response message: %v", resp.Msg)
	if resp.Err != nil {
		log.Error("Replication failed for '%v': %v", head, resp.Err)
	}
}

func Replicate(api api.APIService, store api.RemoteStore, interval time.Duration, topics []crdt.IPFSPath) (chan<- interface{}, <-chan error) {
	stopch := make(chan interface{}, 1)
	errch := make(chan error, len(topics))

	if interval == 0 {
		interval = __DEFAULT_REPLICATE_INTERVAL
	}

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
const __DEFAULT_REPLICATE_INTERVAL = time.Second * 10
