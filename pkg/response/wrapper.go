// costa92/go-protoc/go-protoc-acef4f0ceb39155a2d2db028033d358440154368/pkg/response/wrapper.go

package response

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/proto"
)

// Wrapper 是统一的响应包装结构
type Wrapper struct {
	// Status 表示请求状态，"success" 或 "error"
	Status string `json:"status"`
	// Code 表示 HTTP 状态码
	Code int `json:"code"`
	// Message 包含响应的消息说明
	Message string `json:"message"`
	// Data 包含响应的具体数据
	Data interface{} `json:"data,omitempty"`
	// Error 包含详细错误信息，仅在开发环境中返回
	Error interface{} `json:"error,omitempty"`
}

// CustomMarshaler implements the runtime.Marshaler interface
type CustomMarshaler struct {
	runtime.Marshaler
}

// Marshal wraps the successful proto.Message into the Wrapper struct.
func (c *CustomMarshaler) Marshal(v interface{}) ([]byte, error) {
	// Check if the value is an error, if so, let the error handler deal with it.
	if _, ok := v.(error); ok {
		return c.Marshaler.Marshal(v)
	}

	// Wrap the successful response
	wrappedSuccess := Wrapper{
		Status:  "success",
		Data:    v, // v is the original gRPC response message
		Message: "Request completed successfully",
		Code:    http.StatusOK,
	}

	return json.Marshal(wrappedSuccess)
}

// ForwardResponseMessage is a standard function signature in gRPC-Gateway v2
// for intercepting and customizing the response.
// NOTE: This implementation is kept for reference, but the primary wrapping
// is handled by the Marshal method above for broader compatibility.
func ForwardResponseMessage(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	// This function demonstrates an alternative way to wrap responses.
	// However, using a custom Marshaler is often more straightforward.
	wrappedSuccess := Wrapper{
		Status:  "success",
		Data:    resp,
		Message: "Request completed successfully",
		Code:    http.StatusOK,
	}

	for _, o := range opts {
		if err := o(ctx, w, resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", marshaler.ContentType(wrappedSuccess))
	buf, err := marshaler.Marshal(wrappedSuccess)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buf); err != nil {
		log.Errorf("Failed to write response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// NewSuccessResponse 创建一个成功响应
func NewSuccessResponse(data interface{}, message string) *Wrapper {
	if message == "" {
		message = "请求成功"
	}

	return &Wrapper{
		Status:  "success",
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建一个错误响应
func NewErrorResponse(code int, message string, err error) *Wrapper {
	resp := &Wrapper{
		Status:  "error",
		Code:    code,
		Message: message,
	}

	// 仅在开发环境中包含详细错误信息
	if err != nil && isDevEnvironment() {
		resp.Error = err.Error()
	}

	return resp
}

// NewBadRequestResponse 创建400错误响应
func NewBadRequestResponse(message string, err error) *Wrapper {
	return NewErrorResponse(http.StatusBadRequest, message, err)
}

// NewUnauthorizedResponse 创建401错误响应
func NewUnauthorizedResponse(message string, err error) *Wrapper {
	return NewErrorResponse(http.StatusUnauthorized, message, err)
}

// NewForbiddenResponse 创建403错误响应
func NewForbiddenResponse(message string, err error) *Wrapper {
	return NewErrorResponse(http.StatusForbidden, message, err)
}

// NewNotFoundResponse 创建404错误响应
func NewNotFoundResponse(message string, err error) *Wrapper {
	return NewErrorResponse(http.StatusNotFound, message, err)
}

// NewInternalServerErrorResponse 创建500错误响应
func NewInternalServerErrorResponse(message string, err error) *Wrapper {
	return NewErrorResponse(http.StatusInternalServerError, message, err)
}

// isDevEnvironment 检查是否为开发环境
func isDevEnvironment() bool {
	env := getEnv("ENV", "development")
	return env == "development" || env == "dev"
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := defaultValue
	if envValue, exists := os.LookupEnv(key); exists {
		value = envValue
	}
	return value
}
