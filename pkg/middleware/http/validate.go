package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/costa92/go-protoc/pkg/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

// ValidationMiddleware 创建一个验证中间件
func ValidationMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "validate", true)
			r = r.WithContext(ctx)

			rw := &customResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			if rw.status >= http.StatusBadRequest {
				errorMsg := rw.body
				if len(errorMsg) > 0 {
					st := status.FromContextError(ctx.Err())
					if st.Code() == 3 { // InvalidArgument
						var validationError struct {
							Message string `json:"message"`
							Field   string `json:"field"`
						}
						if err := json.Unmarshal(errorMsg, &validationError); err == nil {
							customErr := errors.ErrValidation.WithDetails(map[string]string{
								"field": validationError.Field,
								"error": validationError.Message,
							})
							errors.WriteJSON(w, customErr)
							return
						}
						// 如果无法解析详细错误信息，返回通用验证错误
						errors.WriteJSON(w, errors.ErrValidation)
						return
					}
				}
			}
		})
	}
}

// customResponseWriter 包装 http.ResponseWriter 记录状态码和响应体
type customResponseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

// WriteHeader 实现 http.ResponseWriter 接口
func (rw *customResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *customResponseWriter) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

// sendErrorResponse 发送格式化的错误响应
func sendErrorResponse(w http.ResponseWriter, statusCode int, message []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

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
		mux := runtime.NewServeMux()
		runtime.DefaultHTTPErrorHandler(context.Background(), mux, &runtime.JSONPb{}, w, &http.Request{}, errResp.Err)
	} else {
		// 回退到原始消息
		customErr := errors.NewError(10100, string(message))
		errors.WriteJSON(w, customErr)
	}
}
