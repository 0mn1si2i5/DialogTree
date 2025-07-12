// Path: ./main.go

package main

import (
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/router/cli_router"
)

func main() {
	global.Config = core.ReadConf(true)
	cli_router.Run()
}
