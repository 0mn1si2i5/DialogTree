// Path: ./router/gin_router/enter.go

package gin_router

import (
	"dialogTree/core"
	"dialogTree/global"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Run() {
	core.Init()
	gin.SetMode(global.Config.System.GinMode) // 设置 gin 模式，对应 settings.yaml 中的 gin_mode

	router := gin.Default()
	router.Static("/uploads", "uploads") // 配置静态路由访问上传文件

	routerGroup := router.Group("/api")

	AiRouter(routerGroup)

	addr := global.Config.System.Addr()
	logrus.Infof("gin running on: %s", addr)
	err := router.Run(addr)
	if err != nil {
		logrus.Fatalln("gin run error: ", err)
	}
}
