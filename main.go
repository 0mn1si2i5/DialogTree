// Path: ./main.go

package main

import (
	"dialogTree/common/cres"
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/router/cli_router"
)

func main() {
	global.Config = core.ReadConf(true)
	cres.SetAgentLabel()
	cli_router.Run()
}
