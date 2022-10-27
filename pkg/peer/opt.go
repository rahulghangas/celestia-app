package peer

import (
	"go.uber.org/zap"
	"time"
)

var (
	DefaultAlpha         = 5
	DefaultTimeout       = time.Second
	DefaultGossipTimeout = 3 * time.Second
)

type SyncerOptions struct {
	Alpha         int
	WiggleTimeout time.Duration
}

func DefaultSyncerOptions() SyncerOptions {
	return SyncerOptions{
		Alpha:         DefaultAlpha,
		WiggleTimeout: DefaultTimeout,
	}
}

func (opts SyncerOptions) WithAlpha(alpha int) SyncerOptions {
	opts.Alpha = alpha
	return opts
}

func (opts SyncerOptions) WithWiggleTimeout(timeout time.Duration) SyncerOptions {
	opts.WiggleTimeout = timeout
	return opts
}

type GossiperOptions struct {
	Alpha   int
	Timeout time.Duration
}

func DefaultGossiperOptions() GossiperOptions {
	return GossiperOptions{
		Alpha:   DefaultAlpha,
		Timeout: DefaultGossipTimeout,
	}
}

func (opts GossiperOptions) WithAlpha(alpha int) GossiperOptions {
	opts.Alpha = alpha
	return opts
}

func (opts GossiperOptions) WithTimeout(timeout time.Duration) GossiperOptions {
	opts.Timeout = timeout
	return opts
}

type DiscoveryOptions struct {
	Logger           *zap.Logger
	Alpha            int
	MaxExpectedPeers int
	PingTimePeriod   time.Duration
}

func DefaultDiscoveryOptions() DiscoveryOptions {
	return DiscoveryOptions{
		Alpha:            DefaultAlpha,
		MaxExpectedPeers: DefaultAlpha,
		PingTimePeriod:   DefaultTimeout,
	}
}

func (opts DiscoveryOptions) WithAlpha(alpha int) DiscoveryOptions {
	opts.Alpha = alpha
	return opts
}

func (opts DiscoveryOptions) WithMaxExpectedPeers(max int) DiscoveryOptions {
	opts.MaxExpectedPeers = max
	return opts
}

func (opts DiscoveryOptions) WithPingTimePeriod(period time.Duration) DiscoveryOptions {
	opts.PingTimePeriod = period
	return opts
}

type Options struct {
	SyncerOptions
	GossiperOptions
	DiscoveryOptions
}

func DefaultOptions() Options {
	return Options{
		SyncerOptions:    DefaultSyncerOptions(),
		GossiperOptions:  DefaultGossiperOptions(),
		DiscoveryOptions: DefaultDiscoveryOptions(),
	}
}

func (opts Options) WithSyncerOptions(syncerOptions SyncerOptions) Options {
	opts.SyncerOptions = syncerOptions
	return opts
}

func (opts Options) WithGossiperOptions(gossiperOptions GossiperOptions) Options {
	opts.GossiperOptions = gossiperOptions
	return opts
}

func (opts Options) WithDiscoveryOptions(discoveryOptions DiscoveryOptions) Options {
	opts.DiscoveryOptions = discoveryOptions
	return opts
}
