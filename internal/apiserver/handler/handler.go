package handler

import (
	"github.com/costa92/go-protoc/v2/internal/apiserver/biz"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewHandler, wire.Bind(new(v1.ApiServerServer), new(*Handler)))

type Handler struct {
	v1.UnimplementedApiServerServer

	biz biz.IBiz
}

var _ v1.ApiServerServer = (*Handler)(nil)

// NewHandler creates a new instance of *Handler.
func NewHandler(biz biz.IBiz) *Handler {
	return &Handler{biz: biz}
}
