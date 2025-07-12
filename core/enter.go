// Path: ./core/enter.go

package core

import (
	"dialogTree/global"
)

//func QuickChat() {
//	if len(os.Args) == 1 || (len(os.Args) > 1 && (os.Args[1] == "chitchat" || os.Args[1] == "c")) || len(os.Args) > 1 && (os.Args[1] == "你好") {
//		fmt.Println("== QuickChat ==")
//		global.Config = ReadConf(true)
//		cli_router.Run()
//		os.Exit(1)
//	}
//}

func CoreInit() {
	InitLogrus()               // 初始化日志文件
	global.DB = InitDB()       // 连接 mysql
	global.Redis = InitRedis() // 连接 redis
}
