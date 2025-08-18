package logger

import (
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// LevelAwareCore wraps a zapcore.Core to conditionally exclude function information
// based on the log level. By default, it excludes function information for Info level logs.
type LevelAwareCore struct {
	zapcore.Core
	excludeAtLevels map[zapcore.Level]bool
	originalConfig  zapcore.EncoderConfig
	encoderType     string
}

// NewLevelAwareCore creates a new core that can exclude function information based on log level.
func NewLevelAwareCore(core zapcore.Core, config zapcore.EncoderConfig, encoderType string) zapcore.Core {
	return &LevelAwareCore{
		Core:            core,
		originalConfig:  config,
		encoderType:     encoderType,
		excludeAtLevels: map[zapcore.Level]bool{
			zapcore.InfoLevel: true, // 在 info 级别时禁用 func 字段
		},
	}
}

// With adds structured context to the Core.
func (c *LevelAwareCore) With(fields []zapcore.Field) zapcore.Core {
	return &LevelAwareCore{
		Core:            c.Core.With(fields),
		originalConfig:  c.originalConfig,
		encoderType:     c.encoderType,
		excludeAtLevels: c.excludeAtLevels,
	}
}

// Check determines whether the supplied Entry should be logged.
func (c *LevelAwareCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write serializes the Entry and any Fields supplied at the log site and writes them to their destination.
func (c *LevelAwareCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// If this level should exclude function information, temporarily modify the entry
	if c.excludeAtLevels[ent.Level] {
		// Create a modified entry without caller info to exclude function
		modifiedEntry := ent
		modifiedEntry.Caller = zapcore.EntryCaller{}
	
		return c.Core.Write(modifiedEntry, fields)
	}
	
	// For other levels, write normally
	return c.Core.Write(ent, fields)
}

// NewLevelAwareEncoder creates an encoder that excludes function info for info level logs.
func NewLevelAwareEncoder(config zapcore.EncoderConfig, encoderType string) zapcore.Encoder {
	// Create two encoder configs - one with function key and one without
	configWithFunction := config
	configWithoutFunction := config
	configWithoutFunction.FunctionKey = zapcore.OmitKey

	var encoderWithFunc, encoderWithoutFunc zapcore.Encoder
	if encoderType == "console" {
		encoderWithFunc = zapcore.NewConsoleEncoder(configWithFunction)
		encoderWithoutFunc = zapcore.NewConsoleEncoder(configWithoutFunction)
	} else {
		encoderWithFunc = zapcore.NewJSONEncoder(configWithFunction)
		encoderWithoutFunc = zapcore.NewJSONEncoder(configWithoutFunction)
	}

	return &LevelAwareEncoder{
		encoderWithFunc:    encoderWithFunc,
		encoderWithoutFunc: encoderWithoutFunc,
	}
}

// LevelAwareEncoder conditionally excludes function information based on log level
type LevelAwareEncoder struct {
	encoderWithFunc    zapcore.Encoder
	encoderWithoutFunc zapcore.Encoder
}

// Clone creates a copy of the encoder
func (e *LevelAwareEncoder) Clone() zapcore.Encoder {
	return &LevelAwareEncoder{
		encoderWithFunc:    e.encoderWithFunc.Clone(),
		encoderWithoutFunc: e.encoderWithoutFunc.Clone(),
	}
}

// EncodeEntry encodes the log entry, conditionally excluding function information for info level
func (e *LevelAwareEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// For info level, use encoder without function
	if entry.Level == zapcore.InfoLevel {
		return e.encoderWithoutFunc.EncodeEntry(entry, fields)
	}
	// For all other levels, use encoder with function
	return e.encoderWithFunc.EncodeEntry(entry, fields)
}

// Implement ObjectEncoder interface by delegating to the appropriate encoder
func (e *LevelAwareEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return e.encoderWithFunc.AddArray(key, marshaler)
}

func (e *LevelAwareEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return e.encoderWithFunc.AddObject(key, marshaler)
}

func (e *LevelAwareEncoder) AddBinary(key string, value []byte) {
	e.encoderWithFunc.AddBinary(key, value)
}

func (e *LevelAwareEncoder) AddByteString(key string, value []byte) {
	e.encoderWithFunc.AddByteString(key, value)
}

func (e *LevelAwareEncoder) AddBool(key string, value bool) {
	e.encoderWithFunc.AddBool(key, value)
}

func (e *LevelAwareEncoder) AddComplex128(key string, value complex128) {
	e.encoderWithFunc.AddComplex128(key, value)
}

func (e *LevelAwareEncoder) AddComplex64(key string, value complex64) {
	e.encoderWithFunc.AddComplex64(key, value)
}

func (e *LevelAwareEncoder) AddDuration(key string, value time.Duration) {
	e.encoderWithFunc.AddDuration(key, value)
}

func (e *LevelAwareEncoder) AddFloat64(key string, value float64) {
	e.encoderWithFunc.AddFloat64(key, value)
}

func (e *LevelAwareEncoder) AddFloat32(key string, value float32) {
	e.encoderWithFunc.AddFloat32(key, value)
}

func (e *LevelAwareEncoder) AddInt(key string, value int) {
	e.encoderWithFunc.AddInt(key, value)
}

func (e *LevelAwareEncoder) AddInt64(key string, value int64) {
	e.encoderWithFunc.AddInt64(key, value)
}

func (e *LevelAwareEncoder) AddInt32(key string, value int32) {
	e.encoderWithFunc.AddInt32(key, value)
}

func (e *LevelAwareEncoder) AddInt16(key string, value int16) {
	e.encoderWithFunc.AddInt16(key, value)
}

func (e *LevelAwareEncoder) AddInt8(key string, value int8) {
	e.encoderWithFunc.AddInt8(key, value)
}

func (e *LevelAwareEncoder) AddString(key, value string) {
	e.encoderWithFunc.AddString(key, value)
}

func (e *LevelAwareEncoder) AddTime(key string, value time.Time) {
	e.encoderWithFunc.AddTime(key, value)
}

func (e *LevelAwareEncoder) AddUint(key string, value uint) {
	e.encoderWithFunc.AddUint(key, value)
}

func (e *LevelAwareEncoder) AddUint64(key string, value uint64) {
	e.encoderWithFunc.AddUint64(key, value)
}

func (e *LevelAwareEncoder) AddUint32(key string, value uint32) {
	e.encoderWithFunc.AddUint32(key, value)
}

func (e *LevelAwareEncoder) AddUint16(key string, value uint16) {
	e.encoderWithFunc.AddUint16(key, value)
}

func (e *LevelAwareEncoder) AddUint8(key string, value uint8) {
	e.encoderWithFunc.AddUint8(key, value)
}

func (e *LevelAwareEncoder) AddUintptr(key string, value uintptr) {
	e.encoderWithFunc.AddUintptr(key, value)
}

func (e *LevelAwareEncoder) AddReflected(key string, value interface{}) error {
	return e.encoderWithFunc.AddReflected(key, value)
}

func (e *LevelAwareEncoder) OpenNamespace(key string) {
	e.encoderWithFunc.OpenNamespace(key)
}