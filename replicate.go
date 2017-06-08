package godless

import "time"

type replicator struct {
	peers    []RemoteStoreAddress
	interval time.Duration
	api      APIPeerService
}

func (p2p replicator) replicate(stopch <-chan interface{}) {
	if len(p2p.peers) == 0 {
		return
	}

	timer := time.NewTicker(p2p.interval)

	// The API supports replicating one peer at a time.
	// The approach taken here is to round robin the peers.
	for i := 0; ; i++ {
		if i >= len(p2p.peers) {
			i = 0
		}

		peer := p2p.peers[i]

	LOOP:
		for {
			select {
			case <-stopch:
				return
			case <-timer.C:
				p2p.replicatePeer(peer)
				break LOOP
			}
		}
	}
}

func (p2p replicator) replicatePeer(peer RemoteStoreAddress) {
	respch, err := p2p.api.Replicate(peer)

	if err != nil {
		logerr("Replication failed (early API error): %v", err)
	}

	resp := <-respch

	if resp.Err == nil {
		loginfo("Replicated peer '%v': '%v'", resp.Msg)
	} else {
		logerr("Replication error (%v): %v", resp.Msg, resp.Err)
	}
}

func Replicate(api APIPeerService, interval time.Duration, peers []RemoteStoreAddress) (chan<- interface{}, error) {
	stopch := make(chan interface{})

	p2p := replicator{
		peers:    peers,
		interval: interval,
		api:      api,
	}

	go p2p.replicate(stopch)

	return stopch, nil
}
