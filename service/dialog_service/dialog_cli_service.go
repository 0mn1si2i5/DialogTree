// Path: ./service/dialog_service/dialog_cli_service.go

package dialog_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"dialogTree/service/ai_service/chat_anywhere"
	"fmt"
	"strings"
)

// CliDialogService CLI 对话服务
type CliDialogService struct{}

var CliDialogServiceInstance = &CliDialogService{}

// GetSessionList 获取会话列表（CLI用）
func (s *CliDialogService) GetSessionList() ([]models.SessionModel, error) {
	var sessions []models.SessionModel
	err := global.DB.Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}

// GetRecentSession 获取最近的会话
func (s *CliDialogService) GetRecentSession() (*models.SessionModel, error) {
	var session models.SessionModel
	err := global.DB.Order("updated_at DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// CreateQuickSession 创建快速会话（用于 CLI 快速对话）
func (s *CliDialogService) CreateQuickSession(title string) (*models.SessionModel, error) {
	session := models.SessionModel{
		Tittle:     title,
		Summary:    "",
		CategoryID: 1, // 默认分类
	}

	err := global.DB.Create(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetSessionDialogTree 获取会话的对话树（CLI展示用）
func (s *CliDialogService) GetSessionDialogTree(sessionID int64) ([]models.DialogModel, error) {
	var dialogs []models.DialogModel
	err := global.DB.Where("session_id = ?", sessionID).
		Preload("ConversationModels").
		Order("created_at ASC").
		Find(&dialogs).Error

	return dialogs, err
}

// StartDialogChat 开始对话聊天（CLI 交互式）
func (s *CliDialogService) StartDialogChat(sessionID int64, parentDialogID *int64) error {
	for {
		fmt.Print("你: ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil || input == "exit" || input == "quit" {
			fmt.Println("退出对话。")
			break
		}

		if strings.TrimSpace(input) == "" {
			continue
		}

		// 处理对话
		err = s.ProcessDialogMessage(sessionID, parentDialogID, input)
		if err != nil {
			fmt.Printf("处理消息失败: %v\n", err)
			continue
		}
	}

	return nil
}

// ProcessDialogMessage 处理单条对话消息
func (s *CliDialogService) ProcessDialogMessage(sessionID int64, parentDialogID *int64, content string) error {
	// 构建上下文
	context, err := BuildDialogContext(sessionID, parentDialogID, content)
	if err != nil {
		return fmt.Errorf("构建上下文失败: %v", err)
	}

	// 准备完整消息
	fullMessage := context + "\n\n用户问题：" + content

	// 调用AI
	msgChan, sumChan, err := chat_anywhere.ChatStreamSum(fullMessage)
	if err != nil {
		return fmt.Errorf("AI服务调用失败: %v", err)
	}

	// 显示AI回复
	fmt.Print("AI: ")
	var fullAnswer strings.Builder
	for chunk := range msgChan {
		fmt.Print(chunk)
		fullAnswer.WriteString(chunk)
	}
	fmt.Println() // 换行

	// 获取摘要
	var summary string
	for s := range sumChan {
		summary += s
	}

	// 保存对话记录
	err = s.SaveDialogRecord(sessionID, parentDialogID, content, fullAnswer.String(), summary)
	if err != nil {
		fmt.Printf("保存对话失败: %v\n", err)
	}

	return nil
}

// SaveDialogRecord 保存对话记录
func (s *CliDialogService) SaveDialogRecord(sessionID int64, parentDialogID *int64, prompt, answer, summaryRaw string) error {
	// 简化的摘要处理
	var title, summary string
	if summaryRaw != "" {
		// 这里可以添加JSON解析逻辑，简化版本直接使用
		title = "CLI对话"
		summary = prompt[:min(100, len(prompt))]
	} else {
		title = "对话"
		summary = prompt[:min(50, len(prompt))]
	}

	// 创建或获取 Dialog
	var dialogID int64
	var isNewSession bool

	if parentDialogID == nil {
		// 创建新的根对话
		dialog := models.DialogModel{
			SessionID: sessionID,
			ParentID:  nil,
		}
		err := global.DB.Create(&dialog).Error
		if err != nil {
			return fmt.Errorf("创建对话节点失败: %v", err)
		}
		dialogID = dialog.ID
		isNewSession = true
	} else {
		// 在指定节点创建子对话
		dialog := models.DialogModel{
			SessionID: sessionID,
			ParentID:  parentDialogID,
		}
		err := global.DB.Create(&dialog).Error
		if err != nil {
			return fmt.Errorf("创建对话节点失败: %v", err)
		}
		dialogID = dialog.ID
	}

	// 创建会话记录
	conversation := models.ConversationModel{
		Prompt:    prompt,
		Answer:    answer,
		SessionID: sessionID,
		DialogID:  dialogID,
		Title:     title,
		Summary:   summary,
		IsStarred: false,
		Comment:   "",
	}

	err := global.DB.Create(&conversation).Error
	if err != nil {
		return fmt.Errorf("创建会话记录失败: %v", err)
	}

	// 如果是新会话的第一条对话，更新会话信息
	if isNewSession {
		updates := map[string]interface{}{
			"tittle":         title,
			"summary":        summary,
			"root_dialog_id": &dialogID,
		}
		err = global.DB.Model(&models.SessionModel{}).Where("id = ?", sessionID).Updates(updates).Error
		if err != nil {
			fmt.Printf("更新会话信息失败: %v\n", err)
		}
	}

	// 异步处理向量化存储
	go func() {
		err := StoreConversationVector(conversation.ID, prompt, answer, summary)
		if err != nil {
			fmt.Printf("向量化存储失败: %v\n", err)
		}
	}()

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
