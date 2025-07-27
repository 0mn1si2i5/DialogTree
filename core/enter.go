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

func Init() {
	InitLogrus()                    // 初始化日志文件
	global.DB = InitDB()            // 连接 mysql
	global.Redis = InitRedis(false) // 连接 redis
}

func InitWithVector() {
	InitLogrus()                    // 初始化日志文件
	global.DB = InitDB()            // 连接 mysql
	global.Redis = InitRedis(false) // 连接 redis

	// 初始化向量服务
	err := InitVector()
	if err != nil {
		panic("向量服务初始化失败: " + err.Error())
	}
}
