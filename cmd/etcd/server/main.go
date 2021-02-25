package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

const service = "test-service"

type helloService struct {
	pb.UnimplementedHelloServer
}

func (h *helloService) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoRequest, error) {
	fmt.Println(req.Message)
	return req, nil
}

func (h *helloService) Stream(req *pb.EchoRequest, srv pb.Hello_StreamServer) error {
	fmt.Println("stream", req.Message)
	i := 1
	for i < 5 {
		if srv.Context().Err() != nil {
			return srv.Context().Err()
		}

		err := srv.Send(req)
		if err != nil {
			return err
		}
		i++
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

func register(ctx context.Context, endpoints []string, self string) error {
	client, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: time.Second * 5})
	if err != nil {
		return err
	}
	resp, err := client.Grant(ctx, 2)
	if err != nil {
		return errors.Wrap(err, "etcd grant")
	}
	_, err = client.Put(ctx, strings.Join([]string{service, self}, "/"), self, clientv3.WithLease(resp.ID))
	if err != nil {
		return errors.Wrap(err, "etcd put")
	}
	respCh, err := client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errors.Wrap(err, "etcd keep alive")
	}

	for range respCh {

	}
	return nil
}

func main() {
	port := flag.String("port", ":8080", "listen port")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := register(ctx, []string{"0.0.0.0:2379"}, "localhost"+*port)
		if err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterHelloServer(s, &helloService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
