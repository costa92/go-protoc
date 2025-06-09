package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/costa92/go-protoc/internal/apiserver"
	"github.com/costa92/go-protoc/pkg/log"
)

func main() {
	// 获取配置文件路径
	configPath := apiserver.GetConfigPath()

	// 创建服务器实例
	server, err := apiserver.NewServer(configPath)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("接收到信号: %s", sig.String())
		cancel()
	}()

	// 启动服务器
	if err := server.Start(ctx); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
