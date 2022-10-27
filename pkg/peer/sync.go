package peer

import (
	"context"
	"github.com/celestiaorg/celestia-app/pkg/channel"
	"github.com/celestiaorg/celestia-app/pkg/transport"
	"github.com/celestiaorg/celestia-app/pkg/wire"
	"sync"
	"time"
)

type pendingContent struct {
	content []byte

	cond *sync.Cond
}

func (pending *pendingContent) wait() <-chan []byte {
	w := make(chan []byte, 1)
	go func() {
		pending.cond.L.Lock()
		for pending.content == nil {
			pending.cond.Wait()
		}
		content := make([]byte, len(pending.content), len(pending.content))
		copy(content, pending.content)
		pending.cond.L.Unlock()
		w <- content
	}()
	return w
}

func (pending *pendingContent) signal(content []byte) {
	pending.cond.L.Lock()
	pending.content = content
	pending.cond.L.Unlock()
	pending.cond.Broadcast()
}

type Syncer struct {
	opts      SyncerOptions
	filter    *channel.SyncFilter
	transport *transport.Transport

	pendingMu *sync.Mutex
	pending   map[string]*pendingContent
}

func NewSyncer(opts SyncerOptions, filter *channel.SyncFilter, transport *transport.Transport) *Syncer {
	return &Syncer{
		opts:      opts,
		filter:    filter,
		transport: transport,

		pendingMu: new(sync.Mutex),
		pending:   make(map[string]*pendingContent, 1024),
	}
}

func (syncer *Syncer) Sync(ctx context.Context, contentID []byte, hint *[]byte) ([]byte, error) {
	syncer.pendingMu.Lock()
	pending, ok := syncer.pending[string(contentID)]
	if !ok {
		pending = &pendingContent{
			content: nil,
			cond:    sync.NewCond(new(sync.Mutex)),
		}
		syncer.pending[string(contentID)] = pending
	}
	syncer.pendingMu.Unlock()

	syncer.filter.Allow(contentID)
	defer func() {
		go func() {
			time.Sleep(syncer.opts.WiggleTimeout)
			syncer.filter.Deny(contentID)
		}()
	}()

	defer func() {
		syncer.pendingMu.Lock()
		delete(syncer.pending, string(contentID))
		syncer.pendingMu.Unlock()
	}()

	if ok {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case content := <-pending.wait():
			return content, nil
		}
	}

	peers := [][]byte{}
	if hint != nil {
		// append hint to peers to list (kind of acts like persistent seeds / peers) but only a one time recommendation
	}

	for i := range peers {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		p := peers[i]
		go func() {
			err := syncer.transport.Send(ctx, p, wire.Msg{
				Version: wire.MsgVersion1,
				Type:    wire.MsgTypePull,
				Data:    contentID,
			})
			if err != nil {
				// log error
			}
		}()
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case content := <-pending.wait():
		return content, nil
	}
}

func (syncer *Syncer) DidReceiveMessage(from []byte, msg wire.Msg) error {
	if msg.Type == wire.MsgTypeSync {
		if syncer.filter.Filter(from, msg) {
			return nil
		}
		syncer.pendingMu.Lock()
		pending, ok := syncer.pending[string(msg.Data)]
		if ok && msg.SyncData != nil {
			pending.signal(msg.SyncData)
		}
		syncer.pendingMu.Unlock()
	}
	return nil
}
