package http

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

// 验证中间件包装 HTTP 请求并在转发到 gRPC 服务前应用验证
// 这个中间件与 grpc-gateway 配合使用
func ValidationMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 参数验证错误由 gRPC 服务端验证拦截器处理
			// 这个中间件只捕获验证错误并将其转换为 HTTP 响应
			ctx := context.WithValue(r.Context(), "validate", true)
			r = r.WithContext(ctx)

			// 捕获后续处理中的错误
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			// 如果已经返回错误，检查是否为验证错误并格式化响应
			if rw.statusCode >= http.StatusBadRequest {
				errorMsg := rw.body
				if len(errorMsg) > 0 {
					st := status.FromContextError(ctx.Err())
					if st.Code() == 3 { // InvalidArgument
						sendErrorResponse(w, http.StatusBadRequest, errorMsg)
						return
					}
				}
			}
		})
	}
}

// responseWriter 包装 http.ResponseWriter 记录状态码和响应体
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// 实现 http.ResponseWriter 接口
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

// sendErrorResponse 发送格式化的错误响应
func sendErrorResponse(w http.ResponseWriter, statusCode int, message []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 尝试解析 gRPC 错误信息格式并将其转换为 HTTP 友好格式
	var grpcError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(message, &grpcError); err == nil && grpcError.Message != "" {
		// 使用 grpc-gateway 错误格式
		errResp := &runtime.HTTPStatusError{
			HTTPStatus: statusCode,
			Err:        status.Errorf(3, grpcError.Message),
		}
		runtime.DefaultHTTPErrorHandler(context.Background(), nil, nil, errResp, w)
	} else {
		// 回退到原始消息
		w.Write(message)
	}
}