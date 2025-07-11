// Path: ./router/gin_router/ai_router.go

package gin_router

import (
	"dialogTree/api"
	"dialogTree/api/ai_api"
	"dialogTree/middleware"
	"github.com/gin-gonic/gin"
)

func AiRouter(rg *gin.RouterGroup) {
	app := api.App.AiApi

	rg.GET("chat", middleware.BindJsonMiddleware[ai_api.AiChatReq], app.NewMessageView)
}
