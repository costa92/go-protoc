package apiserver

import (
	"github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/server"
)

// init registers the v1 error code mapper for apiserver.
func init() {
	// Register HTTP status code mapper for v1 errors
	mapper := server.NewHTTPStatusCodeMapper(v1.ErrorReason_value)
	server.RegisterErrorCodeMapper(mapper)
}