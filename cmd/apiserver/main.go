package main

import (
	_ "go.uber.org/automaxprocs"

	"fmt"
	"os"

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
	return func(basename string) error {
		// 使用 Wire 初始化 API 服务器
		apiServer, err := apiserver.InitializeAPIServer()
		if err != nil {
			fmt.Printf("初始化 API 服务器失败: %v\n", err)
			os.Exit(1)
		}

		// 运行 API 服务器
		return apiServer.Run(opts)
	}
}
