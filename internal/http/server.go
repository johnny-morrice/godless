package http

import (
	"net"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
)

func Serve(laddr string, handler http.Handler) (api.Closer, error) {
	const protocol = "tcp"
	listener, err := net.Listen(protocol, laddr)

	if err != nil {
		return api.Closer{}, errors.Wrap(err, "Serve failed")
	}

	stopch := make(chan struct{})
	wg := &sync.WaitGroup{}
	closer := api.MakeCloser(stopch, wg)

	wg.Add(__SERVER_PROCESS_COUNT)
	go func() {
		defer wg.Done()
		<-stopch
		blocked := listener.Close()

		if blocked == nil {
			log.Info("Listener closed.")
		} else {
			log.Info("Listener closed with: '%v'", blocked.Error())
		}

	}()

	go func() {
		defer wg.Done()

		httpClose := http.Serve(listener, handler)

		log.Info("HTTP server closed: '%v'", httpClose.Error())
	}()

	return closer, nil
}

const __SERVER_PROCESS_COUNT = 2
