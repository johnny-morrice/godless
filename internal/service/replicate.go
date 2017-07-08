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
	API         api.Service
}

type replicator struct {
	topics   []api.PubSubTopic
	interval time.Duration
	api      api.Service
	store    api.RemoteStore
	errch    chan<- error
	stopch   <-chan struct{}
	keyStore api.KeyStore
}

func Replicate(options ReplicateOptions) (api.Closer, <-chan error) {
	stopch := make(chan struct{})
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

	closer := api.MakeCloser(stopch, wg)

	return closer, errch
}

func (p2p replicator) publishAllTopics() {
	if len(p2p.topics) == 0 {
		log.Info("No topics, not publishing.")
		return
	}

	ticker := time.NewTicker(p2p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p2p.stopch:
			log.Info("Stop publishing")
			break
		case <-ticker.C:
			p2p.publishIndex()
		}
	}
}

func (p2p replicator) publishIndex() {
	log.Info("Publishing index...")

	resp, reflectErr := p2p.sendReflectRequest()

	if reflectErr != nil {
		log.Error("Failed to reflect for index address: %s", reflectErr.Error())
		return
	}

	head := crdt.IPFSPath(resp.Path)

	// API should do this for its HEAD.
	privKeys := p2p.keyStore.GetAllPrivateKeys()

	link, signErr := crdt.SignedLink(head, privKeys)

	if signErr != nil {
		log.Error("Failed to sign Index Link (%s): %s", head, signErr.Error())
		return
	}

	publishErr := p2p.store.PublishAddr(link, p2p.topics)

	if publishErr != nil {
		log.Error("Failed to publish index to all topics: %s", publishErr.Error())
		return
	}

	log.Info("Published index at %v", head)
}

func (p2p replicator) sendReflectRequest() (api.Response, error) {
	log.Info("Replicator getting HEAD from API...")
	failResp := api.RESPONSE_FAIL
	failResp.Type = api.API_REFLECT

	respch, err := p2p.api.Call(api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH})

	if err != nil {
		return failResp, errors.Wrap(err, "Reflection failed (Early API failure)")
	}

	resp := <-respch
	log.Info("Reflection API Response message: %s", resp.Msg)

	if resp.Err != nil {
		return resp, resp.Err
	}

	log.Info("Replicator got HEAD at %s", resp.Path)

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
		log.Debug("Dispatching subscription batch")
		request := api.Request{Type: api.API_REPLICATE, Replicate: batch.batch}
		batch.reset()
		go func() {
			batch.p2p.api.Call(request)
		}()
		return
	default:
		break
	}

	log.Debug("Appended Link to subscriptionBatch")
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

	log.Info("Dispatching subscription updates every: %v", p2p.interval)

	go func() {
		for {
			defer batch.stop()
			select {
			case head, present := <-headch:
				if !present {
					return
				}

				batch.update(head)
			case err, present := <-errch:
				if !present {
					break
				}
				log.Info("Subscription error: %s", err.Error())
				return
			case <-p2p.stopch:
				log.Info("Stop subscribing on %s", topic)
				return
			}
		}

	}()
}

const __REPLICATION_PROCESS_COUNT = 2
const __DEFAULT_REPLICATE_INTERVAL = time.Second * 10
