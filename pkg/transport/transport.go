package transport

import (
	"context"
	"github.com/celestiaorg/celestia-app/pkg/wire"
)

type Transport interface {
	// return address book
	Table() interface{}

	Self() []byte

	Host() string
	Port() uint16

	Send(ctx context.Context, remote []byte, msg wire.Msg) error

	Receive(ctx context.Context, receiver func([]byte, wire.Packet) error)
}
