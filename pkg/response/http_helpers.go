package response

import (
	"encoding/json"
	"net/http"
)

// WriteJSON 将数据以JSON格式写入HTTP响应
func WriteJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	// 将数据包装成统一响应格式
	resp := &Wrapper{
		Status:  "success",
		Code:    statusCode,
		Message: "请求成功",
		Data:    data,
	}

	// 错误响应特殊处理
	if statusCode >= 400 {
		resp.Status = "error"
		resp.Message = "请求失败"

		// 如果data是字符串，作为错误消息
		if msg, ok := data.(string); ok {
			resp.Message = msg
			resp.Data = nil
		}
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 序列化并写入响应
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		// 序列化失败时的应急处理
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","code":500,"message":"序列化响应失败"}`))
		return
	}

	w.Write(jsonBytes)
}

// WriteRawData 直接写入原始数据，不包装
func WriteRawData(w http.ResponseWriter, data []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// WriteSuccess 写入成功响应
func WriteSuccess(w http.ResponseWriter, data interface{}, message string) {
	resp := NewSuccessResponse(data, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Code)

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","code":500,"message":"序列化响应失败"}`))
		return
	}

	w.Write(jsonBytes)
}

// WriteError 写入错误响应
func WriteError(w http.ResponseWriter, code int, message string, err error) {
	resp := NewErrorResponse(code, message, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Code)

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","code":500,"message":"序列化响应失败"}`))
		return
	}

	w.Write(jsonBytes)
}

// WriteBadRequest 写入400错误响应
func WriteBadRequest(w http.ResponseWriter, message string, err error) {
	WriteError(w, http.StatusBadRequest, message, err)
}

// WriteUnauthorized 写入401错误响应
func WriteUnauthorized(w http.ResponseWriter, message string, err error) {
	WriteError(w, http.StatusUnauthorized, message, err)
}

// WriteForbidden 写入403错误响应
func WriteForbidden(w http.ResponseWriter, message string, err error) {
	WriteError(w, http.StatusForbidden, message, err)
}

// WriteNotFound 写入404错误响应
func WriteNotFound(w http.ResponseWriter, message string, err error) {
	WriteError(w, http.StatusNotFound, message, err)
}

// WriteInternalServerError 写入500错误响应
func WriteInternalServerError(w http.ResponseWriter, message string, err error) {
	WriteError(w, http.StatusInternalServerError, message, err)
}
