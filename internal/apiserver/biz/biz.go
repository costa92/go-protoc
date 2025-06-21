package biz

import (
	"github.com/costa92/go-protoc/v2/internal/apiserver/store"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

type IBiz interface{}

// NewBiz creates a new biz.
type biz struct {
	store store.IStore
}

// Ensure that biz implements the IBiz.
var _ IBiz = (*biz)(nil)

// NewBiz creates an instance of IBiz.
func NewBiz(store store.IStore) *biz {
	return &biz{store: store}
}
