// Path: ./main.go

package main

import (
	"dialogTree/core"
	"dialogTree/flags"
	"dialogTree/global"
	"dialogTree/router"
)

func main() {
	flags.Parse()
	global.Config = core.ReadConf()
	core.InitLogrus()               // 初始化日志文件
	global.DB = core.InitDB()       // 连接 mysql
	global.Redis = core.InitRedis() // 连接 redis

	flags.Run()
	router.Run()
}
