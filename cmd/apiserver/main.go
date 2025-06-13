package main

import (
	_ "go.uber.org/automaxprocs"

	"github.com/costa92/go-protoc/internal/apiserver"
	"github.com/costa92/go-protoc/internal/apiserver/options"
	"github.com/costa92/go-protoc/pkg/app"
)

func main() {
	// 1. 创建具体的 Options 实例
	opts := options.NewOptions()

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
func run(opts *options.Options) app.RunFunc {
	return func() error {
		cfg, err := opts.Config()
		if err != nil {
			return err
		}
		return Run()
	}
}

func Run(c *apiserver.Config) error {
	// 这里可以添加实际的业务逻辑
	// 比如启动 gRPC 或 HTTP 服务器等
	// 目前只是一个占位符
	return nil
}
