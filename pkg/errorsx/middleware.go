package errorsx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/costa92/go-protoc/v2/pkg/log"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int32             `json:"code"`
	Reason    string            `json:"reason"`
	Message   string            `json:"message"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp int64             `json:"timestamp"`
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
		return nil
	}

	// 转换为 ErrorX
	errorX := FromError(err)
	
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
	
	// 添加请求 ID
	if requestID := getRequestID(ctx); requestID != "" {
		resp.RequestID = requestID
	}
	
	return resp
}

// logError 记录错误日志
func (h *DefaultErrorHandler) logError(ctx context.Context, err *ErrorX) {
	if h.logger == nil {
		return
	}
	
	fields := map[string]any{
		"code":   err.Code,
		"reason": err.Reason,
	}
	
	// 添加元数据
	for k, v := range err.Metadata {
		fields[k] = v
	}
	
	// 添加原始错误
	if err.cause != nil {
		fields["cause"] = err.cause.Error()
	}
	
	// 根据错误级别记录日志
	// 将 map[string]any 转换为 []any 格式
	logFields := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		logFields = append(logFields, k, v)
	}
	
	if err.Code >= 500 {
		h.logger.Errorw(err.cause, err.Message, logFields...)
	} else if err.Code >= 400 {
		h.logger.Warnw(err.Message, logFields...)
	} else {
		h.logger.Infow(err.Message, logFields...)
	}
}

// GinErrorMiddleware Gin 错误处理中间件
func GinErrorMiddleware(handler ErrorHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err
			
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
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// getRequestID 从上下文获取请求 ID
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
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
	handler := NewDefaultErrorHandler(nil)
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