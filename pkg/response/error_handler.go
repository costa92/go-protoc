package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CustomHTTPErrorHandler 是自定义的HTTP错误处理器
func CustomHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// 解析gRPC错误状态
	s := status.Convert(err)

	// 根据gRPC错误代码映射HTTP状态码
	httpStatus := HTTPStatusFromCode(s.Code())

	// 创建错误响应
	errorResp := &Wrapper{
		Status:  "error",
		Code:    httpStatus,
		Message: s.Message(),
	}

	// 处理验证错误的特殊情况
	// if s.Code() == codes.InvalidArgument {
	// }

	// 设置HTTP状态码和内容类型
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// 序列化响应
	respBytes, marshalErr := json.Marshal(errorResp)
	if marshalErr != nil {
		w.Write([]byte(fmt.Sprintf(`{"status":"error","code":500,"message":"序列化错误响应失败: %s"}`, marshalErr.Error())))
		return
	}

	w.Write(respBytes)
}

// HTTPStatusFromCode 将gRPC状态码转换为HTTP状态码
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// SetupHTTPErrorHandler 设置自定义的HTTP错误处理器
func SetupHTTPErrorHandler(mux *runtime.ServeMux) {
	// 使用 WithErrorHandler 设置自定义错误处理器
	runtime.WithErrorHandler(CustomHTTPErrorHandler)(mux)
}
