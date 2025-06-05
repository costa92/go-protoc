package main

import (
	_ "net/http/pprof"

	"github.com/costa92/go-protoc/internal/helloworld"
)

func main() {
	helloworld.RunApp()
}
