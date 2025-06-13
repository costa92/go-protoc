package apiserver

import (
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/options"
)

type Config struct {
	GRPCOptions *options.GRPCOptions `json:"grpc_options"`
	HTTPOptions *options.HTTPOptions `json:"http_options"`
}

// Config 定义了 API 服务器的配置结构体
type completedConfig struct {
	*Config
}

// Complete 方法用于完成配置的初始化
// 它会确保 Config 不为 nil，并返回一个 completedConfig 实例
// 如果 Config 为 nil，则创建一个新的 Config 实例
func (cfg *Config) Complete() completedConfig {
	if cfg == nil {
		cfg = &Config{}
	}
	return completedConfig{Config: cfg}
}

func (c completedConfig) New(stopCh <-chan struct{}) (*Server, error) {
	if c.Config == nil {
		c.Config = &Config{}
	}

	// 确保 app 不为 nil
	if app == nil {
		return nil, app.ErrAppNotInitialized
	}

	s := &Server{
		app: app,
	}

	return s, nil
}

type Server struct {
	app *app.App
}

func (s *Server) Run(opts *Config) error {

	return nil
}
