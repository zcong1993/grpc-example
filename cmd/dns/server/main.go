package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc/metadata"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

type helloService struct {
	pb.UnimplementedHelloServer
}

func (h *helloService) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoRequest, error) {
	fmt.Println(req.Message)
	metadata.AppendToOutgoingContext(ctx, "test", "111")
	md := metadata.MD{
		"aaa": []string{"test"},
	}

	grpc.SendHeader(ctx, md)

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

func main() {
	port := flag.String("port", ":8080", "listen port")
	maxConnectionAge := flag.Bool("maxConnectionAge", false, "If set server keepalive MaxConnectionAge.")
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	opts := make([]grpc.ServerOption, 0)

	if *maxConnectionAge {
		// https://github.com/grpc/grpc/issues/12295
		// dns resolver 刷新周期非常长, 但是重新连接时会刷新 dns
		// 所以设置最大连接存活时间来刷新 dns
		opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: time.Second * 30,
		}))
	}

	s := grpc.NewServer(opts...)
	pb.RegisterHelloServer(s, &helloService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
