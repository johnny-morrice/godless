package http

import (
	"log"

	"net"
	"net/http"

	"github.com/pkg/errors"
)

func Serve(laddr string, handler http.Handler) (chan<- interface{}, error) {
	const protocol = "tcp"
	listener, err := net.Listen(protocol, laddr)

	if err != nil {
		return nil, errors.Wrap(err, "Serve failed")
	}

	closer := make(chan interface{})

	go func() {
		<-closer
		blocked := listener.Close()
		log.Printf("Listener closed with: '%v'", blocked)
	}()

	go func() {
		httpClose := http.Serve(listener, handler)

		log.Printf("HTTP server closed: '%v'", httpClose)
	}()

	return closer, nil
}
