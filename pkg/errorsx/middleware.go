package errorsx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/log"
	"github.com/gin-gonic/gin"
)

// 日志字段对象池，用于优化日志字段分配性能
var logFieldsPool = sync.Pool{
	New: func() interface{} {
		return make([]any, 0, 20) // 预分配容量
	},
}

// 响应对象池，用于优化错误响应分配
var errorResponsePool = sync.Pool{
	New: func() interface{} {
		return &ErrorResponse{}
	},
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int32          `json:"code"`
	Reason    string         `json:"reason"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Timestamp string         `json:"timestamp"`
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(ctx context.Context, err error) *ErrorResponse
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	logger log.Logger
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(logger log.Logger) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		logger: logger,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) *ErrorResponse {
	if err == nil {
		return &ErrorResponse{
			Code:      http.StatusOK,
			Reason:    "OK",
			Message:   "Success",
			Timestamp: getCurrentTimestamp(),
		}
	}

	// 转换为 ErrorX
	errorX := FromError(err)
	if errorX == nil {
		return &ErrorResponse{
			Code:      http.StatusInternalServerError,
			Reason:    "INTERNAL_ERROR",
			Message:   "Internal server error",
			Timestamp: getCurrentTimestamp(),
		}
	}

	// 本地化错误
	localizedErr := LocalizeError(ctx, errorX)

	// 记录错误日志
	h.logError(ctx, localizedErr)

	// 构建响应
	resp := &ErrorResponse{
		Code:      localizedErr.Code,
		Reason:    localizedErr.Reason,
		Message:   localizedErr.Message,
		Metadata:  localizedErr.Metadata,
		Timestamp: getCurrentTimestamp(),
	}

	if localizedErr.RequestID != "" {
		resp.RequestID = localizedErr.RequestID
	}

	return resp
}

// logError 记录错误日志
func (h *DefaultErrorHandler) logError(ctx context.Context, err *ErrorX) {
	if h.logger == nil {
		return
	}

	// 从对象池获取字段映射
	fields := logFieldsPool.Get().([]any)
	defer func() {
		// 重置切片长度但保留容量
		fields = fields[:0]
		logFieldsPool.Put(fields)
	}()

	// 添加基础字段
	fields = append(fields, "code", err.Code, "reason", err.Reason)

	// 添加元数据
	for k, v := range err.Metadata {
		fields = append(fields, k, v)
	}

	// 添加原始错误
	if err.cause != nil {
		fields = append(fields, "cause", err.cause.Error())
	}

	// 根据错误级别记录日志
	if err.Code >= 500 {
		h.logger.Errorw(err.cause, err.Message, fields...)
	} else if err.Code >= 400 {
		h.logger.Warnw(err.Message, fields...)
	} else {
		h.logger.Infow(err.Message, fields...)
	}
}

// GinErrorMiddleware Gin 错误处理中间件
func GinErrorMiddleware(handler ErrorHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取第一个错误
			err := c.Errors[0].Err

			// 处理错误
			resp := handler.HandleError(c.Request.Context(), err)
			if resp != nil {
				c.JSON(int(resp.Code), resp)
				c.Abort()
			}
		}
	}
}

// HTTPErrorMiddleware HTTP 错误处理中间件
func HTTPErrorMiddleware(handler ErrorHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 创建自定义 ResponseWriter 来捕获错误
			wrapper := &responseWrapper{
				ResponseWriter: w,
				handler:        handler,
				request:        r,
			}

			next.ServeHTTP(wrapper, r)
		})
	}
}

// responseWrapper 响应包装器
type responseWrapper struct {
	http.ResponseWriter
	handler ErrorHandler
	request *http.Request
	written bool
}

// WriteError 写入错误响应
func (w *responseWrapper) WriteError(err error) {
	if w.written {
		return
	}

	resp := w.handler.HandleError(w.request.Context(), err)
	if resp == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(resp.Code))

	data, _ := json.Marshal(resp)
	w.Write(data)
	w.written = true
}

// 辅助函数

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// AbortWithError 中止请求并返回错误（Gin 专用）
func AbortWithError(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}

// AbortWithErrorX 中止请求并返回 ErrorX（Gin 专用）
func AbortWithErrorX(c *gin.Context, err *ErrorX) {
	c.Error(err)
	c.Abort()
}

// WriteErrorResponse 直接写入错误响应
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	handler := NewDefaultErrorHandler(log.Default())
	resp := handler.HandleError(r.Context(), err)
	if resp == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(resp.Code))

	data, _ := json.Marshal(resp)
	w.Write(data)
}

// RecoverMiddleware 恢复中间件，处理 panic
func RecoverMiddleware(handler ErrorHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("panic: %v", r)
				}

				// 包装为内部错误
				errorX := Wrap(err, 500, "INTERNAL_ERROR", "Internal server error")

				// 处理错误
				resp := handler.HandleError(c.Request.Context(), errorX)
				if resp != nil {
					c.JSON(int(resp.Code), resp)
				}
				c.Abort()
			}
		}()

		c.Next()
	}
}
