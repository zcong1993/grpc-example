package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/resolver"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

const (
	etcdScheme  = "etcd"
	serviceName = "test-service"
)

func main() {
	stream := flag.Bool("stream", false, "If test stream.")
	flag.Parse()

	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"0.0.0.0:2379"}, DialTimeout: time.Second * 5})
	if err != nil {
		log.Fatal(err)
	}

	resolver.Register(&etcdResolverBuilder{client: client})

	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", etcdScheme, serviceName), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)

	i := 1

	for {
		if !*stream {
			resp, err := c.Echo(context.Background(), &pb.EchoRequest{Message: fmt.Sprintf("test-%d", i)})
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp.Message)
			}
		} else {
			s, err := c.Stream(context.Background(), &pb.EchoRequest{Message: fmt.Sprintf("test-%d", i)})
			if err != nil {
				fmt.Println(err)
			} else {
				for {
					msg, err := s.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Println(msg.Message)
				}
			}
		}
		i++
		time.Sleep(time.Second)
	}
}

type etcdResolverBuilder struct {
	client *clientv3.Client
}

// Build creates a new resolver for the given target.
//
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &etcdResolver{
		client: b.client,
		target: target,
		cc:     cc,
		store:  make(map[string]map[string]struct{}),
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
func (b *etcdResolverBuilder) Scheme() string {
	return etcdScheme
}

type etcdResolver struct {
	client *clientv3.Client
	target resolver.Target
	cc     resolver.ClientConn
	store  map[string]map[string]struct{}
	stopCh chan struct{}
}

func (r *etcdResolver) start(ctx context.Context) error {
	target := r.target.Endpoint
	resp, err := r.client.Get(ctx, target+"/", clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "get init endpoints")
	}

	if _, ok := r.store[target]; !ok {
		r.store[target] = make(map[string]struct{})
	}

	for _, kv := range resp.Kvs {
		r.store[target][string(kv.Value)] = struct{}{}
	}

	r.updateTargetState()

	w := clientv3.NewWatcher(r.client)

	go func() {
		rch := w.Watch(ctx, target+"/", clientv3.WithPrefix())
		for {
			select {
			case <-r.stopCh:
				return
			case wresp := <-rch:
				for _, ev := range wresp.Events {
					switch ev.Type {
					case mvccpb.PUT:
						r.store[target][string(ev.Kv.Value)] = struct{}{}
					case mvccpb.DELETE:
						delete(r.store[target], string(ev.Kv.Key))
					}
				}
				r.updateTargetState()
			}
		}
	}()

	return nil
}

func (r *etcdResolver) updateTargetState() {
	target := r.target.Endpoint
	if _, ok := r.store[target]; !ok {
		return
	}
	addrs := make([]resolver.Address, len(r.store[target]))
	i := 0
	for k := range r.store[target] {
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
