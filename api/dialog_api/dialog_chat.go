// Path: ./api/dialog_api/dialog_chat.go

package dialog_api

import (
	"dialogTree/common/res"
	"dialogTree/global"
	"dialogTree/models"
	"dialogTree/service/ai_service"
	"dialogTree/service/ai_service/chat_anywhere"
	"dialogTree/service/dialog_service"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type NewChatReq struct {
	Content              string `json:"content" binding:"required"`
	SessionID            int64  `json:"sessionId" binding:"required"`
	ParentConversationID *int64 `json:"parentConversationId"` // 可选，指定从哪个conversation继续对话（用于分叉）
}

type ChatResponse struct {
	DialogID       int64  `json:"dialogId"`
	ConversationID int64  `json:"conversationId"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
}

// NewChat 发起新对话
func (DialogApi) NewChat(c *gin.Context) {
	var req NewChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMessage("参数错误", c)
		return
	}

	// 检查会话是否存在
	var session models.SessionModel
	err := global.DB.First(&session, req.SessionID).Error
	if err != nil {
		res.FailWithMessage("会话不存在", c)
		return
	}

	// 构建上下文（短期记忆 + 向量检索）- 现在返回JSON格式
	contextJSON, err := dialog_service.BuildDialogContextFromConversation(req.SessionID, req.ParentConversationID, req.Content)
	if err != nil {
		res.Fail(err, "构建上下文失败", c)
		return
	}

	// 直接使用JSON格式的上下文作为消息
	fullMessage := contextJSON

	// 调用AI进行流式对话
	provider := ai_service.GetDefaultProvider()
	msgChan, sumChan, err := ai_service.ChatStreamSum(fullMessage, provider) // todo
	if err != nil {
		res.Fail(err, "AI服务调用失败", c)
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 流式响应
	var fullAnswer strings.Builder
	var summary string
	done := make(chan bool)

	go func() {
		var summaryBuilder strings.Builder
		for s := range sumChan {
			summaryBuilder.WriteString(s)
		}
		summary = summaryBuilder.String()
		close(done)
	}()

	// 在主goroutine中处理消息流，确保c.Writer有效
	for chunk := range msgChan {
		fullAnswer.WriteString(chunk)
		// 发送SSE数据
		c.SSEvent("message", chunk)
		c.Writer.Flush()
	}

	// 等待摘要处理完成
	<-done

	// 保存对话记录
	go SaveChatRecord(req, fullAnswer.String(), summary)

	// 最后的Flush确保所有缓冲数据都已发送
	c.Writer.Flush()
}

// NewChatSync 同步版本的新对话（用于简单测试）
func (DialogApi) NewChatSync(c *gin.Context) {
	var req NewChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMessage("参数错误", c)
		return
	}

	// 检查会话是否存在
	var session models.SessionModel
	err := global.DB.First(&session, req.SessionID).Error
	if err != nil {
		res.FailWithMessage("会话不存在", c)
		return
	}

	// 构建上下文 - 现在返回JSON格式
	contextJSON, err := dialog_service.BuildDialogContextFromConversation(req.SessionID, req.ParentConversationID, req.Content)
	if err != nil {
		res.Fail(err, "构建上下文失败", c)
		return
	}

	// 直接使用JSON格式的上下文作为消息
	fullMessage := contextJSON

	// 调用AI（简化版，直接返回结果）
	msgChan, sumChan, err := chat_anywhere.ChatStreamSum(fullMessage)
	if err != nil {
		res.Fail(err, "AI服务调用失败", c)
		return
	}

	// 收集完整回答
	var fullAnswer strings.Builder
	for chunk := range msgChan {
		fullAnswer.WriteString(chunk)
	}

	var summary string
	for s := range sumChan {
		summary += s
	}

	// 保存对话记录
	response, err := SaveChatRecord(req, fullAnswer.String(), summary)
	if err != nil {
		res.Fail(err, "保存对话失败", c)
		return
	}

	res.OkWithDetail(response, "对话成功", c)
}

type summarizeType struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// SaveChatRecord 保存对话记录的辅助函数
func SaveChatRecord(req NewChatReq, answer, summaryRaw string) (*ChatResponse, error) {
	var s summarizeType
	// AI现在返回纯文本摘要，不再需要解析JSON
	s.Summary = summaryRaw
	// 标题可以从请求内容中截取，或者设置为默认值
	s.Title = req.Content[:min(100, len(req.Content))]

	var dialogID int64
	var isNewSession bool

	if req.ParentConversationID == nil {
		// 没有指定父conversation，在会话根部创建新的对话分支
		dialog := models.DialogModel{
			SessionID: req.SessionID,
			ParentID:  nil,
		}
		err := global.DB.Create(&dialog).Error
		if err != nil {
			return nil, fmt.Errorf("创建对话节点失败: %v", err)
		}
		dialogID = dialog.ID
		isNewSession = true
	} else {
		// 指定了父conversation，需要检查是否分叉
		parentConv := &models.ConversationModel{}
		err := global.DB.First(parentConv, *req.ParentConversationID).Error
		if err != nil {
			return nil, fmt.Errorf("找不到父conversation: %v", err)
		}

		// 检查是否需要分叉
		needsBranching, err := dialog_service.CheckIfBranchingByConversation(*req.ParentConversationID)
		if err != nil {
			return nil, fmt.Errorf("检查分叉失败: %v", err)
		}

		if needsBranching {
			// 需要分叉：创建分叉dialogs
			newDialogID, _, err := dialog_service.CreateBranchingDialogs(req.SessionID, *req.ParentConversationID, parentConv.DialogID)
			if err != nil {
				return nil, fmt.Errorf("创建分叉失败: %v", err)
			}
			dialogID = newDialogID
		} else {
			// 不需要分叉：直接在当前dialog中添加conversation
			dialogID = parentConv.DialogID
		}
	}

	// 创建会话记录
	conversation := models.ConversationModel{
		Prompt:    req.Content,
		Answer:    answer,
		SessionID: req.SessionID,
		DialogID:  dialogID,
		Title:     s.Title,
		Summary:   s.Summary,
		IsStarred: false,
		Comment:   "",
	}

	err := global.DB.Create(&conversation).Error
	if err != nil {
		return nil, fmt.Errorf("创建会话记录失败: %v", err)
	}

	// 如果是新会话的第一条对话，更新会话信息
	if isNewSession {
		updates := map[string]interface{}{
			"tittle":         s.Title,
			"summary":        s.Summary,
			"root_dialog_id": &dialogID,
		}
		err = global.DB.Model(&models.SessionModel{}).Where("id = ?", req.SessionID).Updates(updates).Error
		if err != nil {
			logrus.Errorf("更新会话信息失败: %v", err)
		}
	}

	// 异步处理向量化存储
	go func() {
		err := dialog_service.StoreConversationVector(conversation.ID, req.Content, answer, s.Summary)
		if err != nil {
			logrus.Errorf("向量化存储失败: %v", err)
		}
	}()

	return &ChatResponse{
		DialogID:       dialogID,
		ConversationID: conversation.ID,
		Title:          s.Title,
		Summary:        s.Summary,
	}, nil
}

// StarConversation 标星/取消标星会话
func (DialogApi) StarConversation(c *gin.Context) {
	conversationIdStr := c.Param("conversationId")
	conversationId, err := strconv.ParseInt(conversationIdStr, 10, 64)
	if err != nil {
		res.FailWithMessage("会话ID无效", c)
		return
	}

	var conversation models.ConversationModel
	err = global.DB.First(&conversation, conversationId).Error
	if err != nil {
		res.FailWithMessage("会话不存在", c)
		return
	}

	// 切换标星状态
	conversation.IsStarred = !conversation.IsStarred
	err = global.DB.Save(&conversation).Error
	if err != nil {
		res.Fail(err, "更新失败", c)
		return
	}

	status := "已标星"
	if !conversation.IsStarred {
		status = "已取消标星"
	}

	res.OkWithDetail(gin.H{
		"isStarred": conversation.IsStarred,
	}, status, c)
}

// UpdateConversationComment 更新会话评论
func (DialogApi) UpdateConversationComment(c *gin.Context) {
	conversationIdStr := c.Param("conversationId")
	conversationId, err := strconv.ParseInt(conversationIdStr, 10, 64)
	if err != nil {
		res.FailWithMessage("会话ID无效", c)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMessage("参数错误", c)
		return
	}

	err = global.DB.Model(&models.ConversationModel{}).
		Where("id = ?", conversationId).
		Update("comment", req.Comment).Error
	if err != nil {
		res.Fail(err, "更新评论失败", c)
		return
	}

	res.OkWithMessage("评论更新成功", c)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
