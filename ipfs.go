package godless

import (
	"fmt"
	"net/http"

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
		Client: defaultHttpClient(),
	}

	return peer
}

func (peer *IpfsPeer) Connect() error {
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)

	if !peer.Shell.IsUp() {
		return fmt.Errorf("IpfsPeer is not up at '%v'", peer.Url)
	}

	return nil
}
