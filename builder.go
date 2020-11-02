package mdnsresolver

import (
	"context"

	"github.com/grandcat/zeroconf"
	"google.golang.org/grpc/resolver"
)

// mdnsBuilder implements the builder interface of grpc resolver
type mdnsBuilder struct {
}

// Build creates a new resolver for the given target.
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *mdnsBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	ti, err := parseResolverTarget(target)
	if err != nil {
		return nil, err
	}

	// DNS address (non-IP).
	ctx, cancel := context.WithCancel(context.Background())
	d := &mdnsResolver{
		ti:                   ti,
		ctx:                  ctx,
		cancel:               cancel,
		cc:                   cc,
		entries:              make(chan *zeroconf.ServiceEntry, 10),
		rn:                   make(chan bool, 1),
		disableServiceConfig: opts.DisableServiceConfig,
	}
	err = d.Init()
	if err != nil {
		return nil, err
	}

	d.wg.Add(2)
	go d.lookup()
	go d.watcher()
	d.ResolveNow(resolver.ResolveNowOptions{})
	return d, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *mdnsBuilder) Scheme() string {
	return "mdns"
}

// NewBuilder creates a mdnsBuilder which is used to factory DNS resolvers.
func NewBuilder() resolver.Builder {
	return &mdnsBuilder{}
}
