// Path: ./router/gin_router/ai_router.go

package gin_router

import (
	"dialogTree/api"
	"github.com/gin-gonic/gin"
)

func AiRouter(rg *gin.RouterGroup) {
	sessionApi := api.App.SessionApi
	dialogApi := api.App.DialogApi
	categoryApi := api.App.CategoryApi

	// 会话管理相关路由
	sessionGroup := rg.Group("/sessions")
	{
		sessionGroup.GET("", sessionApi.GetSessionList)                 // 获取会话列表
		sessionGroup.POST("", sessionApi.CreateSession)                 // 创建新会话
		sessionGroup.GET("/:sessionId/tree", sessionApi.GetSessionTree) // 获取会话对话树
		sessionGroup.DELETE("/:sessionId", sessionApi.DeleteSession)    // 删除会话
	}

	// 对话相关路由
	dialogGroup := rg.Group("/dialog")
	{
		dialogGroup.POST("/chat", dialogApi.NewChat)                                                   // 发起新对话（流式）
		dialogGroup.POST("/chat/sync", dialogApi.NewChatSync)                                          // 发起新对话（同步）
		dialogGroup.PUT("/conversations/:conversationId/star", dialogApi.StarConversation)             // 标星/取消标星
		dialogGroup.PUT("/conversations/:conversationId/comment", dialogApi.UpdateConversationComment) // 更新评论
	}

	categoryGroup := rg.Group("/categories")
	{
		categoryGroup.GET("", categoryApi.GetCategoryList)               // 请求分类列表
		categoryGroup.POST("", categoryApi.CreateCategory)               // 创建新分类
		categoryGroup.PUT("/update", categoryApi.UpdateCategory)         // 修改分类
		categoryGroup.DELETE("/:categoryId", categoryApi.DeleteCategory) // 删除分类
	}
}
