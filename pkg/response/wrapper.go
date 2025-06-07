package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Wrapper 是你期望的统一JSON结构体
type Wrapper struct {
	Status  string      `json:"status"`            // 状态，例如 "success" 或 "error"
	Data    interface{} `json:"data,omitempty"`    // 业务数据
	Message string      `json:"message,omitempty"` // 提示信息
	// 如果需要，还可以添加错误码等字段
	// Code int `json:"code,omitempty"`
}

// CustomMarshaler 实现了 runtime.Marshaler 接口
type CustomMarshaler struct {
	runtime.Marshaler
}

// Marshal 将出站的 proto.Message 包装成我们的 Wrapper 结构体
func (c *CustomMarshaler) Marshal(v interface{}) ([]byte, error) {
	// gRPC-Gateway 在调用 Marshal 时，如果是错误，v 会是 error 类型；
	// 如果是成功，v 会是 proto.Message 类型。
	// 但更可靠的拦截点是 ForwardResponseMessage 和自定义的 ErrorHandler，
	// 因为它们能更清晰地区分成功和失败路径。
	// 此处直接调用原始的 Marshaler，包装逻辑主要放在 ForwardResponseMessage 中。
	return c.Marshaler.Marshal(v)
}

// ForwardResponseMessage 是 gRPC-Gateway v2 中用于拦截成功响应并自定义写入 http.ResponseWriter 的标准方式
func ForwardResponseMessage(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	// 将成功的响应包装起来
	wrappedSuccess := Wrapper{
		Status:  "success",
		Data:    resp, // resp 是原始的 gRPC 响应
		Message: "Request completed successfully",
	}

	// 遵循 gRPC-Gateway 的标准流程，应用 header 等选项
	for _, o := range opts {
		if err := o(ctx, w, resp); err != nil {
			// 如果选项应用失败，记录日志并可能返回一个标准错误
			w.WriteHeader(http.StatusInternalServerError)
			// 此处可以返回一个包装后的错误 JSON
			return
		}
	}

	w.Header().Set("Content-Type", marshaler.ContentType(wrappedSuccess))

	// 使用原始的 marshaler 来序列化我们包装后的结构体
	buf, err := marshaler.Marshal(wrappedSuccess)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buf); err != nil {
		// 记录写入失败的日志
	}
}

// CustomHTTPErrorHandler 是一个自定义的错误处理器
// 它将 gRPC 的错误信息包装成我们期望的统一 JSON 格式
func CustomHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	s := status.Convert(err)
	w.Header().Set("Content-Type", marshaler.ContentType(s.Proto()))

	wrappedErr := Wrapper{
		Status:  "error",
		Message: s.Message(),
		Data:    nil,
	}

	buf, _ := json.Marshal(wrappedErr)
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	w.Write(buf)
}
