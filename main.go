// Path: ./main.go

package main

import (
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/router"
)

func main() {
	core.QuickChat()
	global.Config = core.ReadConf(false)
	core.InitLogrus()               // 初始化日志文件
	global.DB = core.InitDB()       // 连接 mysql
	global.Redis = core.InitRedis() // 连接 redis
	router.Run()
}
