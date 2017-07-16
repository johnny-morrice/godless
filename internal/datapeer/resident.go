package datapeer

import (
	"github.com/johnny-morrice/godless/api"
)

type residentMemoryStorage struct {
}

func MakeResidentMemoryStorage() api.ContentAddressableStorage {
	panic("not implemented")
}

type residentMemoryPubSubBus struct {
}

func MakeResidentMemoryPubSubBus() api.PubSubber {
	panic("not implemented")
}
