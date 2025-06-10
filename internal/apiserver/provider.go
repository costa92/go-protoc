package apiserver

import (
	"github.com/costa92/go-protoc/internal/apiserver/options"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/log"
)

func Run(opts *options.Options) error {
	// 初始化日志
	log.Init(opts.Log)
	defer log.Sync() // 确保在退出前同步日志

	// 创建 GRPC 和 HTTP 服务器
	grpcServer := app.NewGRPCServer(opts.GRPCOptions)
	httpServer := app.NewHTTPServer(opts.HTTPOptions)

	// 获取所有注册的 API 组安装器
	apiGroups := GetAPIGroups()

	// 安装所有 API 组到服务器
	for _, group := range apiGroups {
		if err := group.Install(grpcServer, httpServer); err != nil {
			return err
		}
	}

	// 启动 GRPC 和 HTTP 服务器
	if err := grpcServer.Start(); err != nil {
		return err
	}
	if err := httpServer.Start(); err != nil {
		return err
	}

	return nil

}
