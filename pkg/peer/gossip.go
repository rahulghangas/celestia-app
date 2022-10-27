package peer

import (
	"context"
	"sync"

	"github.com/celestiaorg/celestia-app/pkg/channel"
	"github.com/celestiaorg/celestia-app/pkg/dht"
	"github.com/celestiaorg/celestia-app/pkg/transport"
	"github.com/celestiaorg/celestia-app/pkg/wire"
)

type Gossiper struct {
	opts GossiperOptions

	filter    *channel.SyncFilter
	transport *transport.Transport

	subnetsMu *sync.Mutex
	subnets   map[string][]byte

	resolverMu *sync.RWMutex
	resolver   dht.ContentResolver
}

func NewGossiper(opts GossiperOptions, filter *channel.SyncFilter, transport *transport.Transport) *Gossiper {
	return &Gossiper{
		opts: opts,

		filter:    filter,
		transport: transport,

		subnetsMu: new(sync.Mutex),
		subnets:   make(map[string][]byte, 1024),

		resolverMu: new(sync.RWMutex),
		resolver:   nil,
	}
}

func (g *Gossiper) Resolve(resolver dht.ContentResolver) {
	g.resolverMu.Lock()
	defer g.resolverMu.Unlock()

	g.resolver = resolver
}

func (g *Gossiper) Gossip(ctx context.Context, contentID []byte, subnet []byte) {
	if subnet == nil {
		// TODO: Define and assign default subnet (pick persistent peers)
		subnet = nil
	}

	recipients := [][]byte{}
	if subnet.Equal(&DefaultSubnet) {
		recipients = g.transport.Table().Peers(g.opts.Alpha)
	} else {
		if recipients = g.transport.Table().Subnet(subnet); len(recipients) > g.opts.Alpha {
			recipients = recipients[:g.opts.Alpha]
		}
	}

	msg := wire.Msg{Version: wire.MsgVersion1, To: subnet, Type: wire.MsgTypePush, Data: contentID}
	wg := new(sync.WaitGroup)
	for i := range recipients {
		recipient := recipients[i]
		wg.Add(1)
		go func() {
			defer wg.Done()

			innerContext, cancel := context.WithTimeout(ctx, g.opts.Timeout)
			defer cancel()

			_ = g.transport.Send(innerContext, recipient, msg)
		}()
	}
	wg.Wait()
}

func (g *Gossiper) DidReceiveMessage(from []byte, msg wire.Msg) error {
	switch msg.Type {
	case wire.MsgTypePush:
		g.didReceivePush(from, msg)
	case wire.MsgTypePull:
		g.didReceivePull(from, msg)
	case wire.MsgTypeSync:
		if g.filter.Filter(from, msg) {
			return nil
		}
		g.didReceiveSync(from, msg)
	}
	return nil
}

func (g *Gossiper) didReceivePush(from []byte, msg wire.Msg) {
	if len(msg.Data) == 0 {
		return
	}

	g.resolverMu.RLock()
	if g.resolver == nil {
		g.resolverMu.RUnlock()
		return
	}
	if _, ok := g.resolver.QueryContent(msg.Data); ok {
		g.resolverMu.RUnlock()
		return
	}
	g.resolverMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), g.opts.Timeout)

	g.subnetsMu.Lock()
	g.subnets[string(msg.Data)] = msg.To
	g.subnetsMu.Unlock()

	g.filter.Allow(msg.Data)

	go func() {
		<-ctx.Done()
		cancel()

		g.subnetsMu.Lock()
		delete(g.subnets, string(msg.Data))
		g.subnetsMu.Unlock()

		g.filter.Deny(msg.Data)
	}()

	if err := g.transport.Send(ctx, from, wire.Msg{
		Version: wire.MsgVersion1,
		Type:    wire.MsgTypePull,
		// Define from addr
		To:      nil,
		Data:    msg.Data,
	}); err != nil {
		// TODO: log error
		return
	}
}

func (g *Gossiper) didReceivePull(from []byte, msg wire.Msg) {
	if len(msg.Data) == 0 {
		return
	}

	var content []byte
	var contentOk bool
	func() {
		g.resolverMu.RLock()
		defer g.resolverMu.RUnlock()

		if g.resolver == nil {
			return
		}
		content, contentOk = g.resolver.QueryContent(msg.Data)
	}()
	if !contentOk {
		// TODO: log error
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.opts.Timeout)
	defer cancel()

	if err := g.transport.Send(ctx, from, wire.Msg{
		Version:  wire.MsgVersion1,
		To:       []byte,
		Type:     wire.MsgTypeSync,
		Data:     msg.Data,
		SyncData: content,
	}); err != nil {
		// TODO: log error
	}
	return
}

func (g *Gossiper) didReceiveSync(from []byte, msg wire.Msg) {
	g.resolverMu.RLock()
	if g.resolver == nil {
		g.resolverMu.RUnlock()
		return
	}

	_, alreadySeenContent := g.resolver.QueryContent(msg.Data)
	if alreadySeenContent {
		g.resolverMu.RUnlock()
		return
	}
	if len(msg.Data) == 0 || len(msg.SyncData) == 0 {
		g.resolverMu.RUnlock()
		return
	}

	g.resolver.InsertContent(msg.Data, msg.SyncData)
	g.resolverMu.RUnlock()

	g.subnetsMu.Lock()
	subnet, ok := g.subnets[string(msg.Data)]
	g.subnetsMu.Unlock()

	if !ok {
		// The gossip has taken too long, and the subnet was removed from the
		// map to preserve memory. Gossiping cannot continue.
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.opts.Timeout)
	defer cancel()

	g.Gossip(ctx, msg.Data, subnet)
}
