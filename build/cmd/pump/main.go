package main

import (
	"context"

	"github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
	logger.Infow("pump", "version", "0.0.1")
	logger.WithCtx(context.Background(), "key", "value").Infow("pump", "version", "0.0.1")
}
