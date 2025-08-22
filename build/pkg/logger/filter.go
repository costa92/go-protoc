package logger

import (
	"context"
)

// Filter borrow the idea form https://docs.rs/tracing-subscriber/0.3.18/tracing_subscriber/layer/trait.Filter.html
type Filter interface {
	Enabled(ctx context.Context, meta any) bool

	Register(script string, env any) error

	Reset()

	IsEmpty() bool

	All() []string
}

// Metadata is the metadata of the log record
// idea from https://docs.rs/tracing-core/0.1.32/tracing_core/metadata/struct.Metadata.html
type Metadata struct {
	File string `expr:"file"`
	Line int    `expr:"line"`
	Func string `expr:"func"`
	Uid  uint32 `expr:"uid"`
	Msg  string `expr:"msg"`
}
