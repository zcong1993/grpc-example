package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/resolver"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

const (
	exampleScheme      = "example"
	exampleServiceName = "lb.example.grpc.io"
)

var addrs = []string{"localhost:8080", "localhost:8081"}

func main() {
	stream := flag.Bool("stream", false, "If test stream.")
	flag.Parse()

	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), grpc.WithInsecure())
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

// Following is an example name resolver implementation. Read the name
// resolution example to learn more about it.

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: addrs,
		},
	}
	r.start()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}

func init() {
	resolver.Register(&exampleResolverBuilder{})
}
