// Path: ./router/gin_router/enter.go

package gin_router

import (
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Run() {
	core.InitWithVector()
	gin.SetMode(global.Config.System.GinMode) // 设置 gin 模式，对应 settings.yaml 中的 gin_mode

	router := gin.Default()
	
	// 条件性添加访问日志中间件
	if global.Config.System.EnableAccessLog {
		router.Use(middleware.AccessLogMiddleware())
	}
	
	router.Static("/uploads", "uploads") // 配置静态路由访问上传文件

	routerGroup := router.Group("/api")

	AiRouter(routerGroup)

	addr := global.Config.System.Addr()
	logrus.Infof("gin running with development router")
	logrus.Infof("gin running on: %s", addr)
	err := router.Run(addr)
	if err != nil {
		logrus.Fatalln("gin run error: ", err)
	}
}

func RunWithWeb() {
	core.InitWithVector()
	gin.SetMode(global.Config.System.GinMode)

	router := gin.Default()

	// 条件性添加访问日志中间件
	if global.Config.System.EnableAccessLog {
		router.Use(middleware.AccessLogMiddleware())
	}

	// 静态资源：前端打包后的 JS/CSS 资源
	router.Static("/assets", "./web/assets")

	// 静态资源：上传目录
	router.Static("/uploads", "./uploads")

	// Catch-all: 所有其他路径都尝试从 ./web 中返回文件（排除 /uploads）
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 显式排除 /uploads（由独立路由处理）
		if strings.HasPrefix(path, "/uploads") {
			c.Status(404)
			return
		}

		// 检查 web 下是否存在对应资源
		fullPath := "./web" + path
		if _, err := os.Stat(fullPath); err == nil {
			c.File(fullPath)
			return
		}

		// 默认回退到 index.html（前端路由）
		c.File("./web/index.html")
	})

	// 后端 API
	apiGroup := router.Group("/api")
	AiRouter(apiGroup)

	addr := global.Config.System.Addr()
	logrus.Infof("Gin running at: %s", addr)
	if err := router.Run(addr); err != nil {
		logrus.Fatalln("Gin failed: ", err)
	}
}
