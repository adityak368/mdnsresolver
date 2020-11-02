package mdnsresolver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"

	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

var (
	logger = grpclog.Component("mdns")

	minMDNSResRate = 30 * time.Second
)

// mdnsResolver implements the grpc resolver interface for mdns
type mdnsResolver struct {
	ti       *targetInfo
	ctx      context.Context
	cancel   context.CancelFunc
	cc       resolver.ClientConn
	resolver *zeroconf.Resolver
	// entries channel is used by mdns resolver to get service entries
	entries chan *zeroconf.ServiceEntry
	// wg is used to enforce Close() to return after the watcher() goroutine has finished.
	// Otherwise, data race will be possible. [Race Example] in dns_resolver_test we
	// replace the real lookup functions with mocked ones to facilitate testing.
	// If Close() doesn't wait for watcher() goroutine finishes, race detector sometimes
	// will warns lookup (READ the lookup function pointers) inside watcher() goroutine
	// has data race with replaceNetFunc (WRITE the lookup function pointers).
	wg sync.WaitGroup
	// rb channel is used to queue requests for resolutions. At Max single request so that we dont swamp the network
	rn                   chan bool
	disableServiceConfig bool
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (d *mdnsResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	select {
	case d.rn <- true:
	default:
	}
}

// Close closes the resolver.
func (d *mdnsResolver) Close() {
	d.cancel()
	d.wg.Wait()
}

func (d *mdnsResolver) Init() error {

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces(ifaces))
	if err != nil {
		return err
	}

	d.resolver = resolver

	return nil
}

func (d *mdnsResolver) lookup() {
	defer d.wg.Done()
	func() {
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-d.rn:
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			d.resolver.Lookup(ctx, d.ti.instanceName, d.ti.serviceName, d.ti.domain, d.entries)
			// Sleep to prevent excessive re-resolutions. Incoming resolution requests
			// will be queued in d.rn.
			t := time.NewTimer(minMDNSResRate)
			select {
			case <-t.C:
			case <-d.ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}

func (d *mdnsResolver) watcher() {
	defer d.wg.Done()
	for {
		select {
		case <-d.ctx.Done():
			return
		case entry := <-d.entries:
			if entry != nil {
				// Get all addresses from mdns reply
				addrs := make([]resolver.Address, 0)
				for _, ip := range entry.AddrIPv4 {
					addr := fmt.Sprintf("%s:%d", ip.String(), entry.Port)
					addrs = append(addrs, resolver.Address{Addr: addr})
					logger.Infof("Resolved target to %s", addr)
				}
				state := resolver.State{Addresses: addrs}
				d.cc.UpdateState(state)
			}
		}
	}
}
