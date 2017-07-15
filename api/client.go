package api

type Client interface {
	Send(request Request) (Response, error)
}
