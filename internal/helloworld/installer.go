package helloworld

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/costa92/go-protoc/internal/helloworld/service"
	v1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	v2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

// Install 实现 server.Installer 接口
func (i *APIGroupInstaller) Install(router *mux.Router) error {
	// 注册 v1 的 HTTP 路由
	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/hello", i.handleV1SayHello).Methods(http.MethodPost)
	v1Router.HandleFunc("/hello/{name}", i.handleV1SayHelloAgain).Methods(http.MethodGet)

	// 注册 v2 的 HTTP 路由
	v2Router := router.PathPrefix("/v2").Subrouter()
	v2Router.HandleFunc("/hello", i.handleV2SayHello).Methods(http.MethodPost)
	v2Router.HandleFunc("/hello/{name}", i.handleV2SayHelloAgain).Methods(http.MethodGet)

	// 添加健康检查路由
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return nil
}

// RegisterGRPC 实现 server.Installer 接口
func (i *APIGroupInstaller) RegisterGRPC(srv *grpc.Server) error {
	// 注册 v1 和 v2 的 gRPC 服务
	v1.RegisterGreeterServer(srv, i.v1Service)
	v2.RegisterGreeterServer(srv, i.v2Service)
	return nil
}

// V1 HTTP Handlers
func (i *APIGroupInstaller) handleV1SayHello(w http.ResponseWriter, r *http.Request) {
	var req v1.HelloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := i.v1Service.SayHello(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (i *APIGroupInstaller) handleV1SayHelloAgain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	req := &v1.HelloRequest{Name: name}
	resp, err := i.v1Service.SayHelloAgain(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// V2 HTTP Handlers
func (i *APIGroupInstaller) handleV2SayHello(w http.ResponseWriter, r *http.Request) {
	var req v2.HelloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := i.v2Service.SayHello(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (i *APIGroupInstaller) handleV2SayHelloAgain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	req := &v2.HelloRequest{Name: name}
	resp, err := i.v2Service.SayHelloAgain(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
