package channel

import "github.com/celestiaorg/celestia-app/pkg/wire"

type SyncFilter interface {
	Filter([]byte, wire.Msg) bool
	Allow(id []byte)
	Deny(id []byte)
}
