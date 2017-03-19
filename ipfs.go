package godless

import (
	"fmt"
	"net/http"
	"time"

	ipfs "github.com/ipfs/go-ipfs-api"
)


type IpfsPath string

type IpfsPeer struct {
	Url string
	Client *http.Client
	Shell *ipfs.Shell
}

func MakeIpfsPeer(url string) *IpfsPeer {
	peer := &IpfsPeer{
		Url: url,
		Client: &http.Client{},
	}

	peer.Client.Timeout = time.Duration(__DEFAULT_TIMEOUT)

	return peer
}

func (peer *IpfsPeer) Connect() error {
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)

	if !peer.Shell.IsUp() {
		return fmt.Errorf("IpfsPeer is not up at '%v'", peer.Url)
	}

	return nil
}

const __NS_2_S = 1000000000
const __DEFAULT_TIMEOUT = 20 * __NS_2_S
