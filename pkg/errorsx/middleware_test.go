package errorsx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestErrorResponse(t *testing.T) {
	// 测试错误响应结构
	resp := &errorsx.ErrorResponse{
		Code:      400,
		Reason:    "VALIDATION_ERROR",
		Message:   "Validation failed",
		Metadata:  map[string]any{"field": "email"},
		RequestID: "req-123",
		Timestamp: "2023-01-01T00:00:00Z",
	}
	
	assert.Equal(t, int32(400), resp.Code)
	assert.Equal(t, "VALIDATION_ERROR", resp.Reason)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Equal(t, "email", resp.Metadata["field"])
	assert.Equal(t, "req-123", resp.RequestID)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.Timestamp)
}

func TestDefaultErrorHandler_HandleError(t *testing.T) {
	// 创建默认错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 测试处理 ErrorX 错误
	errorX := errorsx.New(404, "NOT_FOUND", "Resource not found")
	errorX = errorX.AddMetadata("resource", "user")
	errorX = errorX.WithRequestID("req-456")
	
	ctx := context.Background()
	resp := handler.HandleError(ctx, errorX)
	
	assert.Equal(t, int32(404), resp.Code)
	assert.Equal(t, "NOT_FOUND", resp.Reason)
	assert.Equal(t, "Resource not found", resp.Message)
	assert.Equal(t, "user", resp.Metadata["resource"])
	assert.Equal(t, "req-456", resp.RequestID)
	assert.NotEmpty(t, resp.Timestamp)
}

func TestDefaultErrorHandler_HandleStandardError(t *testing.T) {
	// 创建默认错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 测试处理标准错误
	stdErr := fmt.Errorf("database connection failed")
	ctx := context.Background()
	resp := handler.HandleError(ctx, stdErr)
	
	assert.Equal(t, errorsx.UnknownCode, resp.Code)
	assert.Equal(t, errorsx.UnknownReason, resp.Reason)
	assert.Equal(t, "database connection failed", resp.Message)
	assert.NotEmpty(t, resp.Timestamp)
}

func TestDefaultErrorHandler_HandleNilError(t *testing.T) {
	// 创建默认错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 测试处理 nil 错误
	ctx := context.Background()
	resp := handler.HandleError(ctx, nil)
	
	assert.Equal(t, int32(200), resp.Code)
	assert.Equal(t, "OK", resp.Reason)
	assert.Equal(t, "Success", resp.Message)
	assert.NotEmpty(t, resp.Timestamp)
}

func TestGinErrorMiddleware(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建 Gin 引擎
	router := gin.New()
	router.Use(errorsx.GinErrorMiddleware(handler))
	
	// 添加测试路由
	router.GET("/error", func(c *gin.Context) {
		err := errorsx.New(400, "VALIDATION_ERROR", "Invalid input")
		c.Error(err)
		c.Abort()
	})
	
	router.GET("/success", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 测试错误情况
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var resp errorsx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int32(400), resp.Code)
	assert.Equal(t, "VALIDATION_ERROR", resp.Reason)
	assert.Equal(t, "Invalid input", resp.Message)
	
	// 测试成功情况
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, 200, w2.Code)
	assert.Contains(t, w2.Body.String(), "success")
}

func TestHTTPErrorMiddleware(t *testing.T) {
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建测试处理器
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := errorsx.New(500, "INTERNAL_ERROR", "Something went wrong")
		// 将错误存储在上下文中
		ctx := context.WithValue(r.Context(), "error", err)
		r = r.WithContext(ctx)
		
		// 模拟错误处理
		if err, ok := r.Context().Value("error").(error); ok {
			resp := handler.HandleError(r.Context(), err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(int(resp.Code))
			json.NewEncoder(w).Encode(resp)
		}
	})
	
	// 包装中间件
	wrappedHandler := errorsx.HTTPErrorMiddleware(handler)(testHandler)
	
	// 测试请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	wrappedHandler.ServeHTTP(w, req)
	
	assert.Equal(t, 500, w.Code)
	
	var resp errorsx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int32(500), resp.Code)
	assert.Equal(t, "INTERNAL_ERROR", resp.Reason)
	assert.Equal(t, "Something went wrong", resp.Message)
}

func TestRecoverMiddleware(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建 Gin 引擎
	router := gin.New()
	router.Use(errorsx.RecoverMiddleware(handler))
	
	// 添加会 panic 的路由
	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})
	
	// 测试 panic 恢复
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 500, w.Code)
	
	var resp errorsx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int32(500), resp.Code)
	assert.Contains(t, resp.Message, "Internal server error")
}

func TestGinErrorMiddleware_MultipleErrors(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建 Gin 引擎
	router := gin.New()
	router.Use(errorsx.GinErrorMiddleware(handler))
	
	// 添加会产生多个错误的路由
	router.GET("/multiple-errors", func(c *gin.Context) {
		// 添加多个错误
		err1 := errorsx.New(400, "ERROR_1", "First error")
		err2 := errorsx.New(401, "ERROR_2", "Second error")
		
		c.Error(err1)
		c.Error(err2)
		c.Abort()
	})
	
	// 测试多个错误（应该返回第一个错误）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multiple-errors", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code) // 应该返回第一个错误的状态码
	
	var resp errorsx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int32(400), resp.Code)
	assert.Equal(t, "ERROR_1", resp.Reason)
	assert.Equal(t, "First error", resp.Message)
}

func TestGinErrorMiddleware_NoErrors(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建 Gin 引擎
	router := gin.New()
	router.Use(errorsx.GinErrorMiddleware(handler))
	
	// 添加正常的路由
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 测试正常情况（不应该被中间件处理）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/normal", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.NotContains(t, w.Body.String(), "error")
}

func TestCustomErrorHandler(t *testing.T) {
	// 创建自定义错误处理器
	customHandler := &CustomErrorHandler{}
	
	// 测试自定义处理
	errorX := errorsx.New(404, "NOT_FOUND", "Resource not found")
	ctx := context.Background()
	resp := customHandler.HandleError(ctx, errorX)
	
	assert.Equal(t, int32(404), resp.Code)
	assert.Equal(t, "NOT_FOUND", resp.Reason)
	assert.Equal(t, "[CUSTOM] Resource not found", resp.Message) // 自定义前缀
	assert.NotEmpty(t, resp.Timestamp)
}

// 自定义错误处理器用于测试
type CustomErrorHandler struct{}

func (h *CustomErrorHandler) HandleError(ctx context.Context, err error) *errorsx.ErrorResponse {
	// 使用默认处理器处理
	defaultHandler := &errorsx.DefaultErrorHandler{}
	resp := defaultHandler.HandleError(ctx, err)
	
	// 添加自定义前缀
	resp.Message = "[CUSTOM] " + resp.Message
	
	return resp
}

func TestErrorMiddleware_Integration(t *testing.T) {
	// 集成测试：测试完整的错误处理流程
	gin.SetMode(gin.TestMode)
	
	// 创建错误处理器
	handler := &errorsx.DefaultErrorHandler{}
	
	// 创建 Gin 引擎
	router := gin.New()
	router.Use(errorsx.RecoverMiddleware(handler))
	router.Use(errorsx.GinErrorMiddleware(handler))
	
	// 添加各种测试路由
	router.GET("/validation-error", func(c *gin.Context) {
		err := errorsx.BadRequest().
			WithReason("VALIDATION_FAILED").
			WithMessage("Email is required").
			AddMetadata("field", "email").
			Build()
		c.Error(err)
		c.Abort()
	})
	
	router.GET("/not-found", func(c *gin.Context) {
		err := errorsx.NotFound().
			WithReason("USER_NOT_FOUND").
			WithMessage("User with ID 123 not found").
			AddMetadata("user_id", "123").
			Build()
		c.Error(err)
		c.Abort()
	})
	
	router.GET("/internal-error", func(c *gin.Context) {
		panic("database connection failed")
	})
	
	// 测试验证错误
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/validation-error", nil)
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, 400, w1.Code)
	var resp1 errorsx.ErrorResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	assert.Equal(t, "VALIDATION_FAILED", resp1.Reason)
	assert.Equal(t, "email", resp1.Metadata["field"])
	
	// 测试未找到错误
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/not-found", nil)
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, 404, w2.Code)
	var resp2 errorsx.ErrorResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.Equal(t, "USER_NOT_FOUND", resp2.Reason)
	assert.Equal(t, "123", resp2.Metadata["user_id"])
	
	// 测试内部错误（panic 恢复）
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/internal-error", nil)
	router.ServeHTTP(w3, req3)
	
	assert.Equal(t, 500, w3.Code)
	var resp3 errorsx.ErrorResponse
	json.Unmarshal(w3.Body.Bytes(), &resp3)
	assert.Contains(t, resp3.Message, "Internal server error")
}