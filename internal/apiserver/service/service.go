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

func NewGreeterV2Server() helloworldv2.GreeterServer {
	return &GreeterV2Server{}
}

func NewGreeterV1Server() helloworldv1.GreeterServer {
	return &GreeterV1Server{}
}

func (s *GreeterV1Server) SayHello(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello " + req.GetName()}, nil
}

func (s *GreeterV1Server) SayHelloAgain(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	return &helloworldv1.HelloReply{Message: "V1: Hello again " + req.GetName()}, nil
}

func (s *GreeterV2Server) SayHello(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello " + req.GetName()}, nil
}

func (s *GreeterV2Server) SayHelloAgain(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	return &helloworldv2.HelloReply{Message: "V2: Hello again " + req.GetName()}, nil
}
