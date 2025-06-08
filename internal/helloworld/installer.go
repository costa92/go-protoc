package helloworld

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/costa92/go-protoc/internal/helloworld/service"
	v1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	v2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/costa92/go-protoc/pkg/response"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

const (
	apiPrefix    = "/api"
	swaggerPath  = "/swagger"
	swaggerDoc   = "/swagger/doc.yaml"
	swaggerIndex = "/swagger/index.html"
)

// APIGroupInstaller 实现了 server.Installer 接口
type APIGroupInstaller struct {
	logger    *zap.Logger
	v1Service v1.GreeterServer
	v2Service v2.GreeterServer
}

// NewInstaller 创建一个新的 APIGroupInstaller
func NewInstaller(logger *zap.Logger) *APIGroupInstaller {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &APIGroupInstaller{
		logger:    logger,
		v1Service: service.NewGreeterV1Server(),
		v2Service: service.NewGreeterV2Server(),
	}
}

// handleSwaggerDoc serves the swagger.yaml file, dynamically adding the "/api" prefix to all paths.
func (i *APIGroupInstaller) handleSwaggerDoc(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("api/swagger/swagger.yaml")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("swagger.yaml not found"))
		i.logger.Error("failed to read swagger.yaml", zap.Error(err))
		return
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(fileBytes, &data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to parse swagger.yaml"))
		i.logger.Error("failed to unmarshal swagger.yaml", zap.Error(err))
		return
	}

	if paths, ok := data["paths"].(map[string]interface{}); ok {
		newPaths := make(map[string]interface{})
		for key, value := range paths {
			newPaths[apiPrefix+key] = value
		}
		data["paths"] = newPaths
	}

	modifiedBytes, err := yaml.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to generate modified swagger.yaml"))
		i.logger.Error("failed to marshal modified swagger.yaml", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Write(modifiedBytes)
}

// Install 实现 server.Installer 接口
func (i *APIGroupInstaller) Install(router *mux.Router) error {
	router.HandleFunc(swaggerDoc, i.handleSwaggerDoc)

	router.PathPrefix(swaggerPath + "/").Handler(httpSwagger.Handler(
		httpSwagger.URL(swaggerDoc),
	))
	i.logger.Info("registered swagger ui handler", zap.String("path", swaggerIndex))

	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(response.CustomHTTPErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &response.CustomMarshaler{
			Marshaler: &runtime.JSONPb{},
		}),
	)

	err := v1.RegisterGreeterHandlerServer(context.Background(), gwmux, i.v1Service)
	if err != nil {
		return fmt.Errorf("failed to register v1 handler: %w", err)
	}

	err = v2.RegisterGreeterHandlerServer(context.Background(), gwmux, i.v2Service)
	if err != nil {
		return fmt.Errorf("failed to register v2 handler: %w", err)
	}

	router.PathPrefix(apiPrefix + "/").Handler(http.StripPrefix(apiPrefix, gwmux))
	return nil
}

// RegisterGRPC 实现 server.Installer 接口
func (i *APIGroupInstaller) RegisterGRPC(srv *grpc.Server) error {
	v1.RegisterGreeterServer(srv, i.v1Service)
	v2.RegisterGreeterServer(srv, i.v2Service)
	return nil
}
