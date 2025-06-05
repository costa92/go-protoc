package service

import (
	"context"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

type GreeterV1Server struct {
	helloworldv1.UnimplementedGreeterServer
}

type GreeterV2Server struct {
	helloworldv2.UnimplementedGreeterServer
}

func NewGreeterV2Server() *GreeterV2Server {
	return &GreeterV2Server{}
}

func NewGreeterV1Server() *GreeterV1Server {
	return &GreeterV1Server{}
}

type GreeterServer struct {
	GreeterV1Server
	GreeterV2Server
}

func NewGreeterServer() *GreeterServer {
	return &GreeterServer{
		GreeterV1Server: *NewGreeterV1Server(),
		GreeterV2Server: *NewGreeterV2Server(),
	}
}

func (s *GreeterV1Server) SayHello(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello, " + req.GetName()}, nil
}

func (s *GreeterV1Server) SayHelloAgain(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello, " + req.GetName()}, nil
}

func (s *GreeterV2Server) SayHello(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello, " + req.GetName()}, nil
}

func (s *GreeterV2Server) SayHelloAgain(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello, " + req.GetName()}, nil
}
