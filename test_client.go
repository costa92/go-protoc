package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// 使用命令行参数获取要请求的URL
	url := "http://localhost:9090/v1/hello/world"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	// 创建一个HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	// 打印状态码和响应内容
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头: %v\n", resp.Header)
	fmt.Printf("响应体: %s\n", body)
}
