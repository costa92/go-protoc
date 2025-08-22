package server

import (
	"context"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	consulapi "github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// The purpose of defining the AppConfig is to demonstrate the usage of wire.Struct.
type KratosAppConfig struct {
	ID        string
	Name      string
	Version   string
	Metadata  map[string]string
	Registrar registry.Registrar
}

// The purpose of defining the AppConfig is to demonstrate the usage of wire.Struct.
type KratosServer struct {
	kapp *kratos.App
}

func NewKratosServer(cfg KratosAppConfig, servers ...transport.Server) (*KratosServer, error) {
	kapp := kratos.New(
		kratos.ID(cfg.ID+"."+cfg.Name),
		kratos.Name(cfg.Name),
		kratos.Version(cfg.Version),
		kratos.Metadata(cfg.Metadata),
		kratos.Logger(NewKratosLogger(cfg.ID, cfg.Name, cfg.Version)),
		kratos.Registrar(cfg.Registrar),
		kratos.Server(servers...),
	)

	return &KratosServer{kapp: kapp}, nil
}

func (s *KratosServer) RunOrDie() {
	logger.Infow("Start to listening the incoming requests", "protocol", "kratos")
	if err := s.kapp.Run(); err != nil {
		logger.Fatalw("Failed to serve kratos application", "err", err)
	}
}

func (s *KratosServer) GracefulStop(ctx context.Context) {
	logger.Infow("Gracefully stop kratos application")
	if err := s.kapp.Stop(); err != nil {
		logger.Errorw("Failed to gracefully shutdown kratos application", "error", err)
	}
}

func NewKratosLogger(id, name, version string) krtlog.Logger {
	// 使用新的 logger 包创建 Kratos 适配器
	return logger.NewKratosLogger(logger.GetDefaultLogger(), id, name, version)
}

// NewKratosLoggerWithGeneric creates a Kratos-compatible logger using the generic logger package.
func NewKratosLoggerWithGeneric(l logger.Logger, id, name, version string) krtlog.Logger {
	return logger.NewKratosLogger(l, id, name, version)
}

// NewKratosLoggerFromOptions creates a Kratos-compatible logger using logger options.
func NewKratosLoggerFromOptions(opts *logger.LogsOptions, id, name, version string) krtlog.Logger {
	l, err := logger.NewLogger(opts)
	if err != nil {
		// Fallback to default logger if creation fails
		l = logger.GetDefaultLogger()
	}
	return logger.NewKratosLogger(l, id, name, version)
}

func NewEtcdRegistrar(opts *genericoptions.EtcdOptions) registry.Registrar {
	if opts == nil {
		panic("etcd registrar options must be set.")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   opts.Endpoints,
		DialTimeout: opts.DialTimeout,
		TLS:         opts.TLSOptions.MustTLSConfig(),
		Username:    opts.Username,
		Password:    opts.Password,
	})
	if err != nil {
		panic(err)
	}
	r := etcd.New(client)
	return r
}

func NewConsulRegistrar(opts *genericoptions.ConsulOptions) registry.Registrar {
	if opts == nil {
		panic("consul registrar options must be set.")
	}

	c := consulapi.DefaultConfig()
	c.Address = opts.Addr
	c.Scheme = opts.Scheme
	cli, err := consulapi.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}
