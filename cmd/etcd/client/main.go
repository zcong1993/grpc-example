package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/zcong1993/grpc-example/pkg/etcdresolver"

	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/resolver"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

const defaultService = "test-service"

func testService(service string, stream bool) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", etcdresolver.Scheme, service), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)

	i := 1

	for {
		if !stream {
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

func main() {
	stream := flag.Bool("stream", false, "If test stream.")
	service := flag.String("service", defaultService, "service name")
	flag.Parse()

	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"0.0.0.0:2379"}, DialTimeout: time.Second * 5})
	if err != nil {
		log.Fatal(err)
	}

	b := etcdresolver.NewBuilder(client)

	resolver.Register(b)

	tmpArr := strings.Split(*service, ",")
	for _, s := range tmpArr {
		s := s
		go testService(s, *stream)
	}

	for {
		time.Sleep(time.Second * 5)
		b.DebugStore()
	}
}
