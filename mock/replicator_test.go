package mock_godless

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestReplicateSuccess(t *testing.T) {
	// As design stands, replicator is intrinsically a long running process.
	if testing.Short() {
		t.SkipNow()
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockApi := NewMockService(ctrl)

	const head = crdt.IPFSPath("HEAD")
	const interval = time.Millisecond * 100
	const topic = api.PubSubTopic("Topic")
	topics := []api.PubSubTopic{topic}
	link := crdt.UnsignedLink(head)
	links := []crdt.Link{link}
	replicateRequest := api.Request{Type: api.API_REPLICATE, Replicate: links}
	headRequest := api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH}
	headResponse := api.RESPONSE_REFLECT
	headResponse.Path = head

	replicateRespch := make(chan api.Response, 1)
	headRespch := make(chan api.Response, 1)
	suberrch := make(chan error)
	linkch := make(chan crdt.Link, 1)
	linkch <- link
	replicateRespch <- api.RESPONSE_REPLICATE
	headRespch <- headResponse

	mockApi.EXPECT().Call(replicateRequest).Return(replicateRespch, nil).MinTimes(1)
	mockApi.EXPECT().Call(headRequest).Return(headRespch, nil).MinTimes(1)
	mockStore.EXPECT().PublishAddr(link, topics).Return(nil).MinTimes(1)
	mockStore.EXPECT().SubscribeAddrStream(topic).Return(linkch, suberrch).MinTimes(1)

	keyStore := &crypto.KeyStore{}

	options := service.ReplicateOptions{
		Topics:      topics,
		Interval:    interval,
		KeyStore:    keyStore,
		RemoteStore: mockStore,
		API:         mockApi,
	}

	closer, errch := service.Replicate(options)

	defer tidyReplicator(t, closer, errch)
	defer close(suberrch)

	timeout := time.NewTimer(interval)
	<-timeout.C
}

func TestReplicateApiReplicateFailure(t *testing.T) {
	// As design stands, replicator is intrinsically a long running process.
	if testing.Short() {
		t.SkipNow()
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockApi := NewMockService(ctrl)

	const head = crdt.IPFSPath("HEAD")
	const interval = time.Millisecond * 100
	const topic = api.PubSubTopic("Topic")
	topics := []api.PubSubTopic{topic}
	link := crdt.UnsignedLink(head)
	links := []crdt.Link{link}
	replicateRequest := api.Request{Type: api.API_REPLICATE, Replicate: links}
	headRequest := api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH}
	headResponse := api.RESPONSE_REFLECT
	headResponse.Path = head

	replicateRespch := make(chan api.Response, 1)
	headRespch := make(chan api.Response, 1)
	suberrch := make(chan error)
	linkch := make(chan crdt.Link, 1)
	linkch <- link
	respFail := api.RESPONSE_FAIL
	respFail.Err = expectedError()
	replicateRespch <- respFail
	headRespch <- headResponse

	mockStore.EXPECT().PublishAddr(link, topics).Return(nil).MinTimes(1)
	mockStore.EXPECT().SubscribeAddrStream(topic).Return(linkch, suberrch).MinTimes(1)
	mockApi.EXPECT().Call(replicateRequest).Return(replicateRespch, nil).MinTimes(1)
	mockApi.EXPECT().Call(headRequest).Return(headRespch, nil).MinTimes(1)

	keyStore := &crypto.KeyStore{}

	options := service.ReplicateOptions{
		Topics:      topics,
		Interval:    interval,
		KeyStore:    keyStore,
		RemoteStore: mockStore,
		API:         mockApi,
	}

	closer, errch := service.Replicate(options)

	defer tidyReplicator(t, closer, errch)
	defer close(suberrch)

	timeout := time.NewTimer(interval)
	<-timeout.C
}

func TestReplicateApiHeadFailure(t *testing.T) {
	// As design stands, replicator is intrinsically a long running process.
	if testing.Short() {
		t.SkipNow()
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockApi := NewMockService(ctrl)

	const head = crdt.IPFSPath("HEAD")
	const interval = time.Millisecond * 100
	const topic = api.PubSubTopic("Topic")
	topics := []api.PubSubTopic{topic}
	link := crdt.UnsignedLink(head)
	links := []crdt.Link{link}
	replicateRequest := api.Request{Type: api.API_REPLICATE, Replicate: links}
	headRequest := api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH}
	headResponse := api.RESPONSE_REFLECT
	headResponse.Path = head

	replicateRespch := make(chan api.Response, 1)
	headRespch := make(chan api.Response, 1)
	suberrch := make(chan error)
	linkch := make(chan crdt.Link, 1)
	linkch <- link
	respFail := api.RESPONSE_FAIL
	respFail.Err = expectedError()
	replicateRespch <- api.RESPONSE_REPLICATE
	headRespch <- respFail

	mockStore.EXPECT().SubscribeAddrStream(topic).Return(linkch, suberrch).MinTimes(1)
	mockApi.EXPECT().Call(replicateRequest).Return(replicateRespch, nil).MinTimes(1)
	mockApi.EXPECT().Call(headRequest).Return(headRespch, nil).MinTimes(1)

	keyStore := &crypto.KeyStore{}

	options := service.ReplicateOptions{
		Topics:      topics,
		Interval:    interval,
		KeyStore:    keyStore,
		RemoteStore: mockStore,
		API:         mockApi,
	}

	closer, errch := service.Replicate(options)
	defer tidyReplicator(t, closer, errch)

	timeout := time.NewTimer(interval)
	<-timeout.C
}

func TestReplicatePublishFailure(t *testing.T) {
	// As design stands, replicator is intrinsically a long running process.
	if testing.Short() {
		t.SkipNow()
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockApi := NewMockService(ctrl)

	const head = crdt.IPFSPath("HEAD")
	const interval = time.Millisecond * 100
	const topic = api.PubSubTopic("Topic")
	topics := []api.PubSubTopic{topic}
	link := crdt.UnsignedLink(head)
	links := []crdt.Link{link}
	replicateRequest := api.Request{Type: api.API_REPLICATE, Replicate: links}
	headRequest := api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH}
	headResponse := api.RESPONSE_REFLECT
	headResponse.Path = head

	replicateRespch := make(chan api.Response, 1)
	headRespch := make(chan api.Response, 2)
	suberrch := make(chan error)
	linkch := make(chan crdt.Link, 1)
	linkch <- link
	replicateRespch <- api.RESPONSE_REPLICATE
	headRespch <- headResponse
	headRespch <- headResponse

	mockApi.EXPECT().Call(replicateRequest).Return(replicateRespch, nil).MinTimes(1)
	mockApi.EXPECT().Call(headRequest).Return(headRespch, nil).MinTimes(1)

	mockStore.EXPECT().PublishAddr(link, topics).Return(expectedError()).MinTimes(2)
	mockStore.EXPECT().SubscribeAddrStream(topic).Return(linkch, suberrch).MinTimes(1)

	keyStore := &crypto.KeyStore{}

	options := service.ReplicateOptions{
		Topics:      topics,
		Interval:    interval,
		KeyStore:    keyStore,
		RemoteStore: mockStore,
		API:         mockApi,
	}

	closer, errch := service.Replicate(options)

	defer tidyReplicator(t, closer, errch)
	defer close(suberrch)

	timeout := time.NewTimer(interval * 5)
	<-timeout.C
}

func TestReplicateSubscribeFailure(t *testing.T) {
	// As design stands, replicator is intrinsically a long running process.
	if testing.Short() {
		t.SkipNow()
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockApi := NewMockService(ctrl)

	const head = crdt.IPFSPath("HEAD")
	const interval = time.Millisecond * 100
	const topic = api.PubSubTopic("Topic")
	topics := []api.PubSubTopic{topic}
	link := crdt.UnsignedLink(head)
	links := []crdt.Link{link}
	replicateRequest := api.Request{Type: api.API_REPLICATE, Replicate: links}
	headRequest := api.Request{Type: api.API_REFLECT, Reflection: api.REFLECT_HEAD_PATH}
	headResponse := api.RESPONSE_REFLECT
	headResponse.Path = head

	replicateRespch := make(chan api.Response, 1)
	headRespch := make(chan api.Response, 1)
	suberrch := make(chan error)
	linkch := make(chan crdt.Link, 1)
	linkch <- link
	replicateRespch <- api.RESPONSE_REPLICATE
	headRespch <- headResponse

	mockApi.EXPECT().Call(replicateRequest).Return(replicateRespch, nil).MinTimes(1)
	mockApi.EXPECT().Call(headRequest).Return(headRespch, nil).MinTimes(1)

	mockStore.EXPECT().PublishAddr(link, topics).Return(nil)
	mockStore.EXPECT().SubscribeAddrStream(topic).Return(linkch, suberrch)

	keyStore := &crypto.KeyStore{}

	options := service.ReplicateOptions{
		Topics:      topics,
		Interval:    interval,
		KeyStore:    keyStore,
		RemoteStore: mockStore,
		API:         mockApi,
	}

	closer, errch := service.Replicate(options)

	defer tidyReplicator(t, closer, errch)
	defer close(suberrch)

	timeout := time.NewTimer(interval)
	<-timeout.C
}

func tidyReplicator(t *testing.T, closer api.Closer, errch <-chan error) {
	closer.Close()

	for err := range errch {
		testutil.AssertNil(t, err)
	}
}
