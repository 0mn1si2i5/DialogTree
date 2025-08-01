// Path: ./router/gin_router/ai_router.go

package gin_router

import (
	"dialogTree/api"
	"dialogTree/middleware"
	"github.com/gin-gonic/gin"
)

func AiRouter(rg *gin.RouterGroup) {
	sessionApi := api.App.SessionApi
	dialogApi := api.App.DialogApi
	categoryApi := api.App.CategoryApi

	// 会话管理相关路由
	sessionGroup := rg.Group("/sessions")
	{
		sessionGroup.GET("", sessionApi.GetSessionList)                                         // 获取会话列表
		sessionGroup.POST("", middleware.DemoMiddleware, sessionApi.CreateSession)              // 创建新会话
		sessionGroup.GET("/:sessionId/tree", sessionApi.GetSessionTree)                         // 获取会话对话树
		sessionGroup.PUT("/:sessionId", middleware.DemoMiddleware, sessionApi.UpdateSession)    // 更新会话信息
		sessionGroup.DELETE("/:sessionId", middleware.DemoMiddleware, sessionApi.DeleteSession) // 删除会话
	}

	// 对话相关路由
	dialogGroup := rg.Group("/dialog")
	{
		dialogGroup.POST("/chat", middleware.DemoMiddleware, dialogApi.NewChat)                                              // 发起新对话（流式）
		dialogGroup.POST("/chat/sync", middleware.DemoMiddleware, dialogApi.NewChatSync)                                     // 发起新对话（同步）
		dialogGroup.GET("/conversations/:conversationId/ancestors", dialogApi.GetAncestors)                                  // 获取祖先对话
		dialogGroup.PUT("/conversations/:conversationId/star", middleware.DemoMiddleware, dialogApi.StarConversation)        // 标星/取消标星
		dialogGroup.PUT("/conversations/comment", middleware.DemoMiddleware, dialogApi.UpdateConversationComment)            // 更新评论
		dialogGroup.PUT("/conversations/title", middleware.DemoMiddleware, dialogApi.UpdateConversationTitle)                // 更新标题
		dialogGroup.DELETE("/conversations/:conversationId", middleware.DemoMiddleware, dialogApi.DeleteConversationComment) // 删除评论
	}

	categoryGroup := rg.Group("/categories")
	{
		categoryGroup.GET("", categoryApi.GetCategoryList)                                          // 请求分类列表
		categoryGroup.POST("", middleware.DemoMiddleware, categoryApi.CreateCategory)               // 创建新分类
		categoryGroup.PUT("/update", middleware.DemoMiddleware, categoryApi.UpdateCategory)         // 修改分类
		categoryGroup.DELETE("/:categoryId", middleware.DemoMiddleware, categoryApi.DeleteCategory) // 删除分类
		categoryGroup.GET("/:categoryId/sessions", sessionApi.GetSessionsByCategory)                // 获取分类下的所有会话
	}
}
