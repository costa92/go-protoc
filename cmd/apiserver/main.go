package main

import (
	"fmt"

	"github.com/costa92/go-protoc/cmd/apiserver/app/options"
	"github.com/costa92/go-protoc/pkg/app"
	_ "go.uber.org/automaxprocs"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

func main() {
	// 1. 创建具体的 Options 实例
	opts := options.NewServerOptions()

	// 2. 创建一个 App 构建器
	//    - 传入应用名称和二进制文件名
	//    - 注入具体的 Options（它满足 CliOptions 接口）
	//    - 注入真正的业务运行逻辑 (通过 Wire 生成)
	application := app.NewApp("API Server", "apiserver",
		app.WithOptions(opts),
		app.WithRunFunc(run(opts)),
	)

	// 3. 运行应用
	application.Run()
}

// run 函数创建了一个闭包，它持有具体的 options 实例
func run(opts *options.ServerOptions) app.RunFunc {
	return func() error {
		cfg, err := opts.Config()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		ctx := genericapiserver.SetupSignalContext()
		server, err := cfg.NewServer(ctx)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		return server.Run(ctx)
	}
}
