package service

import (
	"sync"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type ReplicateOptions struct {
	Topics      []api.PubSubTopic
	Interval    time.Duration
	KeyStore    api.KeyStore
	RemoteStore api.RemoteStore
	API         api.APIService
}

type replicator struct {
	topics   []api.PubSubTopic
	interval time.Duration
	api      api.APIService
	store    api.RemoteStore
	errch    chan<- error
	stopch   <-chan struct{}
	keyStore api.KeyStore
}

func Replicate(options ReplicateOptions) (chan<- struct{}, <-chan error) {
	stopch := make(chan struct{}, 1)
	errch := make(chan error, len(options.Topics))

	interval := options.Interval
	if interval == 0 {
		interval = __DEFAULT_REPLICATE_INTERVAL
	}

	p2p := replicator{
		topics:   options.Topics,
		api:      options.API,
		store:    options.RemoteStore,
		keyStore: options.KeyStore,
		interval: interval,
		errch:    errch,
		stopch:   stopch,
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

	head := crdt.IPFSPath(addr.ReflectResponse.Path)

	privKeys := p2p.keyStore.GetAllPrivateKeys()

	link, signErr := crdt.SignedLink(head, privKeys)

	if signErr != nil {
		log.Error("Failed to sign Index Link (%v): %v", head, signErr.Error())
	}

	publishErr := p2p.store.PublishAddr(link, p2p.topics)

	if publishErr != nil {
		log.Error("Failed to publish index: %v", publishErr.Error())
		return
	}

	log.Info("Published index at %v", head)
}

func (p2p replicator) sendReflectRequest() (api.APIResponse, error) {
	failResp := api.RESPONSE_FAIL
	failResp.Type = api.API_REFLECT

	respch, err := p2p.api.Call(api.APIRequest{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH})

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

type subscriptionBatch struct {
	batch  []crdt.Link
	ticker *time.Ticker
	p2p    replicator
}

func (batch *subscriptionBatch) update(link crdt.Link) {
	batch.batch = append(batch.batch, link)

	select {
	case <-batch.ticker.C:
		request := api.APIRequest{Type: api.API_REPLICATE, Replicate: batch.batch}
		batch.reset()
		go func() {
			batch.p2p.api.Call(request)
		}()
	}
}

func (batch *subscriptionBatch) stop() {
	batch.ticker.Stop()
}

func (batch *subscriptionBatch) reset() {
	batch.batch = []crdt.Link{}
}

func (p2p replicator) subscribeTopic(topic api.PubSubTopic) {
	headch, errch := p2p.store.SubscribeAddrStream(topic)

	batch := &subscriptionBatch{
		ticker: time.NewTicker(p2p.interval),
		p2p:    p2p,
	}

	go func() {
		defer batch.stop()
	LOOP:
		for {
			select {
			case head, present := <-headch:
				if !present {
					break LOOP
				}

				batch.update(head)
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

const __REPLICATION_PROCESS_COUNT = 2
const __DEFAULT_REPLICATE_INTERVAL = time.Second * 10
