package api

import (
	"sync"
)

type Closer struct {
	stopch chan<- struct{}
	wg     *sync.WaitGroup
}

func MakeCloser(stopch chan<- struct{}, wg *sync.WaitGroup) Closer {
	return Closer{
		stopch: stopch,
		wg:     wg,
	}
}

func (closer Closer) Close() {
	close(closer.stopch)
	closer.wg.Wait()
}
