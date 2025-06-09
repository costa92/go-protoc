package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// JSONMarshaler 是标准 JSON 包装器，用于包装响应为统一格式
type JSONMarshaler struct{}

// Marshal 将响应包装为统一的 JSON 格式
func (m *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	// 检查是否已经是 Wrapper 类型
	if wrapper, ok := v.(*Wrapper); ok {
		return json.Marshal(wrapper)
	}

	// 检查是否是错误
	if err, ok := v.(error); ok {
		errWrapper := Wrapper{
			Status:  "error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return json.Marshal(errWrapper)
	}

	// 包装成功响应
	respWrapper := Wrapper{
		Status:  "success",
		Data:    v,
		Message: "请求成功",
		Code:    http.StatusOK,
	}
	return json.Marshal(respWrapper)
}

// Unmarshal 实现 runtime.Marshaler 接口
func (m *JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ContentType 返回内容类型
func (m *JSONMarshaler) ContentType(v interface{}) string {
	return "application/json"
}

// NewDecoder 返回一个新的解码器
func (m *JSONMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder 返回一个新的编码器
func (m *JSONMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// RawDataMarshaler 是原始数据处理器，不做任何包装
type RawDataMarshaler struct{}

// Marshal 直接返回原始数据，无包装
func (m *RawDataMarshaler) Marshal(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	return json.Marshal(v) // 对于非字节数组，使用标准 JSON 编码
}

// Unmarshal 实现 runtime.Marshaler 接口
func (m *RawDataMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ContentType 返回内容类型
func (m *RawDataMarshaler) ContentType(v interface{}) string {
	return "application/octet-stream"
}

// NewDecoder 返回一个新的解码器
func (m *RawDataMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder 返回一个新的编码器
func (m *RawDataMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// FileMarshaler 是文件下载处理器
type FileMarshaler struct{}

// Marshal 处理文件下载
func (m *FileMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch file := v.(type) {
	case *os.File:
		return io.ReadAll(file)
	case []byte:
		return file, nil
	default:
		return nil, errors.New("unsupported file type")
	}
}

// Unmarshal 实现 runtime.Marshaler 接口
func (m *FileMarshaler) Unmarshal(data []byte, v interface{}) error {
	if ptr, ok := v.(*[]byte); ok {
		*ptr = data
		return nil
	}
	return errors.New("unsupported unmarshal type")
}

// ContentType 返回内容类型
func (m *FileMarshaler) ContentType(v interface{}) string {
	return "application/octet-stream"
}

// NewDecoder 返回一个新的解码器
func (m *FileMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder 返回一个新的编码器
func (m *FileMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// SetupMarshalers 配置响应编码器
func SetupMarshalers(mux *runtime.ServeMux) {
	// 注册默认 JSON 包装器
	runtime.WithMarshalerOption(runtime.MIMEWildcard, &JSONMarshaler{})(mux)

	// 注册特殊内容类型的处理器
	runtime.WithMarshalerOption("application/octet-stream", &RawDataMarshaler{})(mux)
	runtime.WithMarshalerOption("application/pdf", &FileMarshaler{})(mux)
	runtime.WithMarshalerOption("image/*", &FileMarshaler{})(mux)
}
