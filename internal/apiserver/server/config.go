package server

import "github.com/costa92/go-protoc/pkg/options"

type Config struct {
	GRPCOptions *options.GRPCOptions `json:"grpc_options"`
	HTTPOptions *options.HTTPOptions `json:"http_options"`
}
