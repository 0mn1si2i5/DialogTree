// Path: ./core/enter.go

package core

import (
	"dialogTree/global"
	"github.com/sirupsen/logrus"
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
	if global.Config.Vector.Enable {
		err := InitVector()
		if err != nil {
			panic("向量服务初始化失败: " + err.Error())
		}
	} else {
		logrus.Warnf("向量数据库未启用，跳过向量数据库初始化...")
	}

	// 初始化IP数据库（仅在启用访问日志时）
	if global.Config.System.EnableAccessLog {
		InitIPDB()
		logrus.Infof("IP数据库初始化成功，访问日志功能已启用")
	} else {
		logrus.Infof("访问日志功能已关闭，跳过IP数据库初始化")
	}
}
