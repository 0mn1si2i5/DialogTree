// Path: ./main.go

package main

import (
	"dialogTree/common/cres"
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/router/cli_router"
	"dialogTree/router/gin_router"
)

func main() {
	global.Config = core.ReadConf(true)
	core.InitWithVector()
	cres.SetAgentLabel()
	cli_router.Run()
	gin_router.Run()
}
