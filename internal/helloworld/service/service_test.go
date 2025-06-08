package service

import (
	"context"
	"testing"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

func TestGreeterV1SayHello(t *testing.T) {
	// 创建服务实例
	server := NewGreeterV1Server()

	// 定义测试用例
	testCases := []struct {
		name     string
		request  *helloworldv1.HelloRequest
		expected string
	}{
		{
			name:     "基本问候",
			request:  &helloworldv1.HelloRequest{Name: "世界"},
			expected: "V1: Hello 世界",
		},
		{
			name:     "空名称",
			request:  &helloworldv1.HelloRequest{Name: ""},
			expected: "V1: Hello ",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 调用服务方法
			resp, err := server.SayHello(context.Background(), tc.request)

			// 验证结果
			if err != nil {
				t.Fatalf("SayHello返回了错误: %v", err)
			}
			if resp.Message != tc.expected {
				t.Errorf("响应不匹配: 期望='%s', 实际='%s'", tc.expected, resp.Message)
			}
		})
	}
}

func TestGreeterV1SayHelloAgain(t *testing.T) {
	// 创建服务实例
	server := NewGreeterV1Server()

	// 定义测试用例
	testCases := []struct {
		name     string
		request  *helloworldv1.HelloRequest
		expected string
	}{
		{
			name:     "再次问候",
			request:  &helloworldv1.HelloRequest{Name: "世界"},
			expected: "V1: Hello again 世界",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 调用服务方法
			resp, err := server.SayHelloAgain(context.Background(), tc.request)

			// 验证结果
			if err != nil {
				t.Fatalf("SayHelloAgain返回了错误: %v", err)
			}
			if resp.Message != tc.expected {
				t.Errorf("响应不匹配: 期望='%s', 实际='%s'", tc.expected, resp.Message)
			}
		})
	}
}

func TestGreeterV2SayHello(t *testing.T) {
	// 创建服务实例
	server := NewGreeterV2Server()

	// 定义测试用例
	testCases := []struct {
		name     string
		request  *helloworldv2.HelloRequest
		expected string
	}{
		{
			name:     "V2基本问候",
			request:  &helloworldv2.HelloRequest{Name: "世界"},
			expected: "V2: Hello 世界",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 调用服务方法
			resp, err := server.SayHello(context.Background(), tc.request)

			// 验证结果
			if err != nil {
				t.Fatalf("SayHello返回了错误: %v", err)
			}
			if resp.Message != tc.expected {
				t.Errorf("响应不匹配: 期望='%s', 实际='%s'", tc.expected, resp.Message)
			}
		})
	}
}

func TestGreeterV2SayHelloAgain(t *testing.T) {
	// 创建服务实例
	server := NewGreeterV2Server()

	// 定义测试用例
	testCases := []struct {
		name     string
		request  *helloworldv2.HelloRequest
		expected string
	}{
		{
			name:     "V2再次问候",
			request:  &helloworldv2.HelloRequest{Name: "世界"},
			expected: "V2: Hello again 世界",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 调用服务方法
			resp, err := server.SayHelloAgain(context.Background(), tc.request)

			// 验证结果
			if err != nil {
				t.Fatalf("SayHelloAgain返回了错误: %v", err)
			}
			if resp.Message != tc.expected {
				t.Errorf("响应不匹配: 期望='%s', 实际='%s'", tc.expected, resp.Message)
			}
		})
	}
}