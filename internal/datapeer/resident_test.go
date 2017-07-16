package datapeer

import (
	"crypto"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestResidentMemoryStorage(t *testing.T) {
	const dataText = "Much data!"

	options := ResidentMemoryStorageOptions{
		Hash: crypto.SHA1,
	}

	storage := MakeResidentMemoryStorage(options)

	data, err := storage.Cat("not present")
	testutil.AssertNil(t, data)
	testutil.AssertNonNil(t, err)

	keyOne, err := storage.Add(strings.NewReader(dataText))
	testutil.AssertNil(t, err)

	data, err = storage.Cat(keyOne)
	testutil.AssertNil(t, err)
	dataBytes, err := ioutil.ReadAll(data)
	testutil.AssertNil(t, err)
	testutil.AssertBytesEqual(t, []byte(dataText), dataBytes)

	keyTwo, err := storage.Add(strings.NewReader(dataText))
	testutil.AssertNil(t, err)
	testutil.AssertEquals(t, "Unexpected hash", keyOne, keyTwo)
}

func TestResidentMemoryPubSub(t *testing.T) {
	const topicA = "Topic A"
	const topicB = "Topic B"
	const dataA = "Data A"
	const dataB = "Data B"
	expectA := []byte(dataA)
	expectB := []byte(dataB)

	pubsubber := MakeResidentMemoryPubSubBus()

	subA1, err := pubsubber.PubSubSubscribe(topicA)
	testutil.AssertNil(t, err)
	subA2, err := pubsubber.PubSubSubscribe(topicA)
	testutil.AssertNil(t, err)
	subB, err := pubsubber.PubSubSubscribe(topicB)
	testutil.AssertNil(t, err)

	pubsubber.PubSubPublish(topicA, dataA)
	pubsubber.PubSubPublish(topicB, dataB)

	recordA1, err := subA1.Next()
	testutil.AssertNil(t, err)
	recordA2, err := subA2.Next()
	testutil.AssertNil(t, err)
	recordB, err := subB.Next()
	testutil.AssertNil(t, err)

	testutil.AssertBytesEqual(t, expectA, recordA1.Data())
	testutil.AssertBytesEqual(t, expectA, recordA2.Data())
	testutil.AssertBytesEqual(t, expectB, recordB.Data())
}
