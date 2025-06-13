package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/costa92/go-protoc/pkg/logger"
	"github.com/costa92/go-protoc/pkg/response"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoValidator 定义 protobuf 消息验证接口
type ProtoValidator interface {
	proto.Message
	Validate() error
}

// ValidationError 自定义验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// protoMessageKey 是用于存储 protobuf 消息的上下文键
type protoMessageKey struct{}

// ProtoValidationMiddleware 创建一个 protobuf 验证中间件
func ProtoValidationMiddleware(v ProtoValidator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
					status.Error(codes.InvalidArgument, "请求体不能为空"))
				return
			}
			defer r.Body.Close()

			// 克隆一个新的验证器实例
			msg := proto.Clone(v).(ProtoValidator)

			// 根据 Content-Type 处理不同类型的请求
			contentType := r.Header.Get("Content-Type")
			switch contentType {
			case "application/json":
				// 解析 JSON 到 protobuf 消息
				unmarshaler := protojson.UnmarshalOptions{
					DiscardUnknown: true,
					AllowPartial:   true,
				}

				// 先尝试解析为原始 JSON
				var rawJson json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&rawJson); err != nil {
					logger.L().Errorf("解析 JSON 失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, "无效的 JSON 格式，请确保：\n1. 不包含注释\n2. 使用双引号包裹字符串\n3. 符合标准的 JSON 语法"))
					return
				}

				// 尝试转换为 protobuf 消息
				if err := unmarshaler.Unmarshal(rawJson, msg); err != nil {
					logger.L().Errorf("JSON 转换到 protobuf 失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, fmt.Sprintf("请求数据格式错误: %v", err)))
					return
				}

				// 执行验证
				if err := msg.Validate(); err != nil {
					logger.L().Errorf("protobuf 验证失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, err.Error()))
					return
				}

			case "application/x-protobuf":
				// 直接解析 protobuf 二进制数据
				data, err := readAll(r.Body)
				if err != nil {
					logger.L().Errorf("读取请求体失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, fmt.Sprintf("读取请求失败: %v", err)))
					return
				}

				if err := proto.Unmarshal(data, msg); err != nil {
					logger.L().Errorf("protobuf 解析失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, fmt.Sprintf("无效的 protobuf 格式: %v", err)))
					return
				}

				// 执行验证
				if err := msg.Validate(); err != nil {
					logger.L().Errorf("protobuf 验证失败: %v", err)
					response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
						status.Error(codes.InvalidArgument, err.Error()))
					return
				}

			default:
				response.CustomHTTPErrorHandler(r.Context(), nil, nil, w, r,
					status.Error(codes.InvalidArgument, "不支持的 Content-Type，请使用 application/json 或 application/x-protobuf"))
				return
			}

			// 将验证后的消息存储到请求上下文中
			ctx := WithProtoMessage(r.Context(), msg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// parseValidationError 解析验证错误，返回结构化的错误信息
func parseValidationError(err error) []map[string]string {
	validationErrors := make([]map[string]string, 0)

	// 尝试将错误转换为字符串并解析
	errStr := err.Error()
	if errStr != "" {
		// 这里可以根据实际的错误格式进行更详细的解析
		// 例如，使用正则表达式匹配字段名和错误消息
		validationError := map[string]string{
			"field":   "request",
			"message": errStr,
		}
		validationErrors = append(validationErrors, validationError)
	}

	return validationErrors
}

// WithProtoMessage 将 protobuf 消息存储到上下文中
func WithProtoMessage(ctx context.Context, msg proto.Message) context.Context {
	return context.WithValue(ctx, protoMessageKey{}, msg)
}

// GetProtoMessage 从上下文中获取 protobuf 消息
func GetProtoMessage(ctx context.Context) proto.Message {
	if msg, ok := ctx.Value(protoMessageKey{}).(proto.Message); ok {
		return msg
	}
	return nil
}

// readAll 读取所有请求体数据
func readAll(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}
