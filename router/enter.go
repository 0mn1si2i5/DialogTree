// Path: ./router/enter.go

package router

import (
	"dialogTree/global"
	"dialogTree/router/cli_router"
	"dialogTree/router/gin_router"
	"github.com/sirupsen/logrus"
	"os"
)

func Run() {
	switch global.Config.System.Mode {
	case "cli":
		cli_router.Run()
	case "web":
		gin_router.Run()
	default:
		logrus.Errorf("%v unknown mode: %s\n", os.Stderr, global.Config.System.Mode)
		os.Exit(1)
	}
}
