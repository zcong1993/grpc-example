package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/zcong1993/grpc-example/pb"
	"google.golang.org/grpc"
)

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

func main() {
	port := flag.String("port", ":8080", "listen port")
	flag.Parse()

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
