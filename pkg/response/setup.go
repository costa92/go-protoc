package response

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Setup 配置响应系统，包括格式化器和错误处理器
func Setup(mux *runtime.ServeMux) {
	// 设置自定义格式化器
	SetupMarshalers(mux)

	// 设置自定义错误处理器
	SetupHTTPErrorHandler(mux)
}
