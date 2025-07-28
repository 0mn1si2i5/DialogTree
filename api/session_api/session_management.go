// Path: ./api/session_api/session_management.go

package session_api

import (
	"dialogTree/common/res"
	"dialogTree/global"
	"dialogTree/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateSessionReq struct {
	Title      string `json:"title" binding:"required"`
	CategoryID int64  `json:"categoryID"`
}

type SessionListResponse struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	CategoryID int64  `json:"categoryID"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type DialogTreeNode struct {
	DialogID      int64              `json:"dialogId"`
	ParentID      *int64             `json:"parentId"`
	Conversations []ConversationInfo `json:"conversations"`
	Children      []*DialogTreeNode  `json:"children"`
}

type ConversationInfo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Prompt    string `json:"prompt"`
	Answer    string `json:"answer"`
	IsStarred bool   `json:"isStarred"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"createdAt"`
}

// GetSessionList 获取会话列表
func (SessionApi) GetSessionList(c *gin.Context) {
	var sessions []models.SessionModel
	err := global.DB.Order("updated_at DESC").Find(&sessions).Error
	if err != nil {
		res.Fail(err, "获取会话列表失败", c)
		return
	}

	var response []SessionListResponse
	for _, session := range sessions {
		response = append(response, SessionListResponse{
			ID:         session.ID,
			Title:      session.Tittle, // 注意：原模型中是 Tittle
			Summary:    session.Summary,
			CategoryID: session.CategoryID,
			CreatedAt:  session.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  session.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	res.OkWithDetail(response, "获取成功", c)
}

// CreateSession 创建新会话
func (SessionApi) CreateSession(c *gin.Context) {
	var req CreateSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMessage("参数错误", c)
		return
	}

	if req.CategoryID == 0 {
		req.CategoryID = 1 // 默认分类
	}

	session := models.SessionModel{
		Tittle:     req.Title,
		Summary:    "",
		CategoryID: req.CategoryID,
	}

	err := global.DB.Create(&session).Error
	if err != nil {
		res.Fail(err, "创建会话失败", c)
		return
	}

	res.OkWithDetail(gin.H{
		"sessionId": session.ID,
		"title":     session.Tittle,
	}, "创建成功", c)
}

// GetSessionTree 获取会话的对话树
func (SessionApi) GetSessionTree(c *gin.Context) {
	sessionIdStr := c.Param("sessionId")
	sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
	if err != nil {
		res.FailWithMessage("会话ID无效", c)
		return
	}

	// 检查会话是否存在
	var session models.SessionModel
	err = global.DB.First(&session, sessionId).Error
	if err != nil {
		res.FailWithMessage("会话不存在", c)
		return
	}

	// 获取所有 dialog
	var dialogs []models.DialogModel
	err = global.DB.Where("session_id = ?", sessionId).
		Preload("ConversationModels").
		Find(&dialogs).Error
	if err != nil {
		res.Fail(err, "获取对话树失败", c)
		return
	}

	// 构建树结构
	tree := buildDialogTree(dialogs)

	res.OkWithDetail(gin.H{
		"sessionId":   sessionId,
		"sessionInfo": session,
		"dialogTree":  tree,
	}, "获取成功", c)
}

// DeleteSession 删除会话
func (SessionApi) DeleteSession(c *gin.Context) {
	sessionIdStr := c.Param("sessionId")
	sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
	if err != nil {
		res.FailWithMessage("会话ID无效", c)
		return
	}

	var m models.SessionModel
	err = global.DB.Find(&m, "id = ?", sessionId).Error
	if err != nil {
		res.Fail(err, "查询出错", c)
		return
	}
	if m.ID == 0 {
		res.FailWithMessage("会话不存在", c)
		return
	}

	err = deleteSessionTransaction(sessionId)
	if err != nil {
		res.Fail(err, "删除会话失败", c)
		return
	}

	res.OkWithMessage("删除成功", c)
}

func deleteSessionTransaction(sessionID int64) error {
	// 开始事务
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除对话（ConversationModel）
	if err := tx.Delete(&models.ConversationModel{}, "session_id = ?", sessionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除对话树（DialogModel）
	if err := tx.Delete(&models.DialogModel{}, "session_id = ?", sessionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除会话（SessionModel）
	if err := tx.Delete(&models.SessionModel{}, "id = ?", sessionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// 构建对话树的辅助函数
func buildDialogTree(dialogs []models.DialogModel) []*DialogTreeNode {
	dialogMap := make(map[int64]*DialogTreeNode)
	var roots []*DialogTreeNode

	// 创建所有节点
	for _, dialog := range dialogs {
		node := &DialogTreeNode{
			DialogID:      dialog.ID,
			ParentID:      dialog.ParentID,
			Conversations: make([]ConversationInfo, 0),
			Children:      make([]*DialogTreeNode, 0),
		}

		// 添加会话信息
		for _, conv := range dialog.ConversationModels {
			node.Conversations = append(node.Conversations, ConversationInfo{
				ID:        conv.ID,
				Title:     conv.Title,
				Summary:   conv.Summary,
				Prompt:    conv.Prompt,
				Answer:    conv.Answer,
				IsStarred: conv.IsStarred,
				Comment:   conv.Comment,
				CreatedAt: conv.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		dialogMap[dialog.ID] = node
	}

	// 构建树结构
	for _, dialog := range dialogs {
		node := dialogMap[dialog.ID]
		if dialog.ParentID == nil {
			// 根节点
			roots = append(roots, node)
		} else {
			// 子节点
			if parent, exists := dialogMap[*dialog.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots
}
