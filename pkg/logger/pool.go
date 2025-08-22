package logger

import (
	"strings"
	"sync"
)

// metadataPool 用于复用 Metadata 结构体，减少内存分配
var metadataPool = sync.Pool{
	New: func() interface{} {
		return &Metadata{}
	},
}

// GetMetadata 从对象池获取 Metadata
func GetMetadata() *Metadata {
	return metadataPool.Get().(*Metadata)
}

// PutMetadata 将 Metadata 放回对象池
func PutMetadata(m *Metadata) {
	// 重置字段
	m.File = ""
	m.Line = 0
	m.Func = ""
	m.Uid = 0
	m.Msg = ""
	metadataPool.Put(m)
}

// stringBuilderPool 用于复用 string builder，减少字符串拼接的内存分配
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// GetStringBuilder 从对象池获取 string builder
func GetStringBuilder() *strings.Builder {
	return stringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder 将 string builder 放回对象池
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	stringBuilderPool.Put(sb)
}
