package etcdresolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

const (
	Scheme = "etcd"
)

type Builder struct {
	client *clientv3.Client
	store  map[string]map[string]struct{}
}

func NewBuilder(client *clientv3.Client) *Builder {
	return &Builder{
		client: client,
		store:  make(map[string]map[string]struct{}),
	}
}

func (b *Builder) DebugStore() {
	fmt.Printf("store %+v\n", b.store)
}

// Build creates a new resolver for the given target.
//
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Printf("call builder %s\n", target.Endpoint)
	b.store[target.Endpoint] = make(map[string]struct{})

	r := &etcdResolver{
		client: b.client,
		target: target,
		cc:     cc,
		store:  b.store[target.Endpoint],
		stopCh: make(chan struct{}),
	}

	err := r.start(context.Background())
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *Builder) Scheme() string {
	return Scheme
}

type etcdResolver struct {
	client *clientv3.Client
	target resolver.Target
	cc     resolver.ClientConn
	store  map[string]struct{}
	stopCh chan struct{}
}

func (r *etcdResolver) start(ctx context.Context) error {
	target := r.target.Endpoint
	resp, err := r.client.Get(ctx, target+"/", clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "get init endpoints")
	}

	for _, kv := range resp.Kvs {
		r.store[string(kv.Value)] = struct{}{}
	}

	r.updateTargetState()

	go func() {
		w := clientv3.NewWatcher(r.client)
		rch := w.Watch(ctx, target+"/", clientv3.WithPrefix())
		for {
			select {
			case <-r.stopCh:
				w.Close()
				return
			case wresp := <-rch:
				for _, ev := range wresp.Events {
					switch ev.Type {
					case mvccpb.PUT:
						r.store[string(ev.Kv.Value)] = struct{}{}
					case mvccpb.DELETE:
						delete(r.store, strings.Replace(string(ev.Kv.Key), target+"/", "", 1))
					}
				}
				r.updateTargetState()
			}
		}
	}()

	return nil
}

func (r *etcdResolver) updateTargetState() {
	addrs := make([]resolver.Address, len(r.store))
	i := 0
	for k := range r.store {
		addrs[i] = resolver.Address{Addr: k}
		i++
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {

}

// Close closes the resolver.
func (r *etcdResolver) Close() {
	close(r.stopCh)
}
