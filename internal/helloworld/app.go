package helloworld

import (
	"context"

	"github.com/costa92/go-protoc/internal/helloworld/service"
	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	// "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RunApp() *app.App {
	logger, _ := zap.NewProduction()
	app := app.NewApp(
		app.WithLogger(logger),
		app.WithServiceRegisterFunc(func(srv *grpc.Server, logger *zap.Logger) {
			helloworldv1.RegisterGreeterServer(srv, service.NewGreeterV1Server())
		}),
		app.WithServiceRegisterFunc(func(srv *grpc.Server, logger *zap.Logger) {
			helloworldv2.RegisterGreeterServer(srv, service.NewGreeterV2Server())
		}),
		app.WithGatewayRegisterFuncs(
			func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
				if err := helloworldv1.RegisterGreeterHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
					return err
				}
				if err := helloworldv2.RegisterGreeterHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
					return err
				}
				return nil
			},
		),
	)
	app.Run()
	return app
}
