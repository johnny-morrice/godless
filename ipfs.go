package godless

import (
	"net/http"
	ipfs "github.com/ipfs/go-ipfs-api"
)


type IpfsPath string

type IpfsPeer struct {
	Url string
	Client *http.Client
	Shell *ipfs.Shell
}

func (peer *IpfsPeer) Connect() {
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)
}
