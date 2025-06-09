package helloworld

import (
	"errors"
	"net/http"
	"time"

	"github.com/costa92/go-protoc/pkg/response"
)

// HealthCheckHandler 返回服务健康状态
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	response.WriteSuccess(w, health, "服务正常")
}

// SimpleErrorHandler 返回错误响应示例
func SimpleErrorHandler(w http.ResponseWriter, r *http.Request) {
	err := errors.New("这是一个示例错误")
	response.WriteError(w, http.StatusInternalServerError, "操作失败", err)
}

// FileDownloadHandler 演示如何返回文件
func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 示例数据
	data := []byte("这是一个示例文件内容")

	// 设置Content-Disposition头，使浏览器下载文件
	w.Header().Set("Content-Disposition", "attachment; filename=example.txt")

	// 使用原始数据写入器返回数据
	response.WriteRawData(w, data, "text/plain")
}
