// Path: ./main.go

package main

import (
	"dialogTree/common/cres"
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/router/cli_router"
	"dialogTree/router/gin_router"
	"os"
)

func main() {
	global.Config = core.ReadConf(true)
	core.InitWithVector()
	cres.SetAgentLabel()

	// 如果没有命令行参数或第一个参数是server,启动HTTP服务器
	if len(os.Args) <= 1 || os.Args[1] == "server" {
		gin_router.Run()
	} else {
		// 否则使用CLI模式
		cli_router.Run()
	}
}
