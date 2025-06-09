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
		log.Infof("启动服务器失败: %v", err)
	}

	// 在服务器停止后，调用 Stop 方法进行完整的清理
	log.Infof("服务器正在关闭，执行清理操作...")
	server.Stop()
	log.Infof("清理完成，程序退出。")

}
