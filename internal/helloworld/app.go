package helloworld

import (
	"github.com/costa92/go-protoc/pkg/app"
)

func RunApp() *app.App {
	app := app.NewApp()
	app.Run()
	return app
}
