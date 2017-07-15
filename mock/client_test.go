package mock_godless

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/h2non/gock"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestClientSend(t *testing.T) {
	const SIZE = 10

	for i := 0; i < SIZE; i++ {
		testClientSendOnce(t, testutil.Rand(), SIZE)
	}
}

func testClientSendOnce(t *testing.T, rand *rand.Rand, size int) {
	defer gock.Off()

	buff := &bytes.Buffer{}
	expected := api.GenResponse(testutil.Rand(), size)
	err := api.EncodeResponse(expected, buff)
	panicOnBadInit(err)

	expect := gock.New(TEST_SERVER_ADDR)
	expect.Post(TEST_COMMAND_ENDPOINT)
	expect.Reply(200)
	expect.ReplyFunc(func(mockReply *gock.Response) {
		mockReply.AddHeader("content-type", "application/octet-stream")
		mockReply.Body(buff)
	})

	options := http.ClientOptions{
		ServerAddr: TEST_SERVER_ADDR,
		Endpoints: http.Endpoints{
			CommandEndpoint: TEST_COMMAND_ENDPOINT,
		},
	}

	client, err := http.MakeClient(options)
	panicOnBadInit(err)

	request := api.GenRequest(testutil.Rand(), size)
	actual, err := client.Send(request)

	testutil.Assert(t, "Unexpected error state", (actual.Err == nil) == (err == nil))
	testutil.Assert(t, "Unexpected api.Response", expected.Equals(actual))
}

const TEST_SERVER_ADDR = "https://example.org"
const TEST_COMMAND_ENDPOINT = "/api"
