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

// ProtoValidationMiddleware 创建一个 protobuf 验证中间件
func ProtoValidationMiddleware(v ProtoValidator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				response.WriteError(w, http.StatusBadRequest, "请求体不能为空", nil)
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

				var rawJson json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&rawJson); err != nil {
					logger.L().Errorf("解析 JSON 失败: %v", err)
					response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("无效的 JSON 格式: %v", err), err)
					return
				}

				if err := unmarshaler.Unmarshal(rawJson, msg); err != nil {
					logger.L().Errorf("JSON 转换到 protobuf 失败: %v", err)
					response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("无效的请求数据格式: %v", err), err)
					return
				}

			case "application/x-protobuf":
				// 直接解析 protobuf 二进制数据
				data, err := readAll(r.Body)
				if err != nil {
					logger.L().Errorf("读取请求体失败: %v", err)
					response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("读取请求失败: %v", err), err)
					return
				}

				if err := proto.Unmarshal(data, msg); err != nil {
					logger.L().Errorf("protobuf 解析失败: %v", err)
					response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("无效的 protobuf 格式: %v", err), err)
					return
				}

			default:
				response.WriteError(w, http.StatusUnsupportedMediaType, "不支持的 Content-Type", nil)
				return
			}

			// 执行验证
			if err := msg.Validate(); err != nil {
				logger.L().Errorf("protobuf 验证失败: %v", err)
				// 处理验证错误，转换为更友好的错误消息
				validationError := parseValidationError(err)
				response.WriteError(w, http.StatusBadRequest, validationError, err)
				return
			}

			// 将验证后的消息存储到请求上下文中
			ctx := WithProtoMessage(r.Context(), msg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// parseValidationError 解析验证错误，返回友好的错误消息
func parseValidationError(err error) string {
	// 这里可以根据实际的验证错误类型进行更详细的解析
	// 例如，如果使用 protoc-gen-validate，可以解析其特定的错误格式
	return fmt.Sprintf("数据验证失败: %v", err)
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

type protoMessageKey struct{}

// readAll 读取所有请求体数据
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
