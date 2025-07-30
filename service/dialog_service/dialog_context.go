// Path: ./service/dialog_service/dialog_context.go

package dialog_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"dialogTree/service/embedding_service"
	"dialogTree/service/vector_service"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// 为了控制每个上下文的最长长度
const maxLength = 1000
const headTailLength = 400

// QAPair 问答对结构
type QAPair struct {
	Q string `json:"Q"`
	A string `json:"A"`
}

// ContextData 上下文数据结构
type ContextData struct {
	Recent  []QAPair `json:"recent"`
	History []QAPair `json:"history"`
	Current string   `json:"current"`
}

// BuildDialogContext 构建对话上下文（短期记忆 + 长期记忆）
func BuildDialogContext(sessionID int64, parentDialogID *int64, currentQuestion string) (string, error) {
	var contextParts []string

	// 1. 构建短期记忆上下文（往上追溯N轮对话）
	shortTermContext, err := buildShortTermContext(sessionID, parentDialogID)
	if err != nil {
		return "", fmt.Errorf("构建短期上下文失败: %v", err)
	}

	if shortTermContext != "" {
		contextParts = append(contextParts, "## 最近对话上下文")
		contextParts = append(contextParts, shortTermContext)
	}

	// 2. 构建长期记忆上下文（向量检索相关历史）
	longTermContext, err := buildLongTermContext(sessionID, currentQuestion)
	if err != nil {
		// 长期记忆检索失败不应该影响整个对话流程，只记录错误
		fmt.Printf("长期记忆检索失败: %v\n", err)
	} else if longTermContext != "" {
		contextParts = append(contextParts, "## 相关历史记忆")
		contextParts = append(contextParts, longTermContext)
	}

	// 3. 如果没有任何上下文，返回空字符串
	if len(contextParts) == 0 {
		return "", nil
	}

	// 4. 添加说明文字
	introduction := "以下是对话的上下文信息，请基于这些信息回答用户的问题：\n"

	return introduction + strings.Join(contextParts, "\n\n"), nil
}

// buildShortTermContext 构建短期记忆上下文
func buildShortTermContext(sessionID int64, parentDialogID *int64) (string, error) {
	contextLayers := global.Config.Ai.ContextLayers
	if contextLayers <= 0 {
		return "", nil
	}

	var conversations []models.ConversationModel
	var err error

	if parentDialogID == nil {
		// 如果没有指定父对话，获取会话中最新的几轮对话（跨Dialog）
		conversations, err = getRecentConversationsAcrossDialogs(sessionID, contextLayers)
	} else {
		// 从指定的对话节点往上追溯
		conversations, err = traceParentConversations(*parentDialogID, contextLayers)
	}

	if err != nil {
		return "", err
	}

	if len(conversations) == 0 {
		return "", nil
	}

	// 按时间正序排列（最早的在前面）
	var contextLines []string
	for i := len(conversations) - 1; i >= 0; i-- {
		conv := conversations[i]
		contextLines = append(contextLines, fmt.Sprintf("Q: %s", conv.Prompt))
		contextLines = append(contextLines, fmt.Sprintf("A: %s", conv.Summary)) // 使用摘要而非完整回答
	}

	return strings.Join(contextLines, "\n"), nil
}

// buildShortTermContextFromConversation 从指定conversation构建短期记忆上下文
// 这个函数专门用于分叉场景，能正确处理跨dialog的追溯
func buildShortTermContextFromConversation(sessionID int64, parentConversationID *int64) (string, error) {
	contextLayers := global.Config.Ai.ContextLayers
	if contextLayers <= 0 {
		return "", nil
	}

	var conversations []models.ConversationModel
	var err error

	if parentConversationID == nil {
		// 如果没有指定父conversation，获取会话中最新的几轮对话（跨Dialog）
		conversations, err = getRecentConversationsAcrossDialogs(sessionID, contextLayers)
	} else {
		// 从指定的conversation开始往上追溯
		conversations, err = traceParentConversationsFromConversation(*parentConversationID, contextLayers)
	}

	if err != nil {
		return "", err
	}

	if len(conversations) == 0 {
		return "", nil
	}

	// 按时间正序排列（最早的在前面）
	var contextLines []string
	for i := len(conversations) - 1; i >= 0; i-- {
		conv := conversations[i]
		contextLines = append(contextLines, fmt.Sprintf("Q: %s", conv.Prompt))
		contextLines = append(contextLines, fmt.Sprintf("A: %s", conv.Summary)) // 使用摘要而非完整回答
	}

	return strings.Join(contextLines, "\n"), nil
}

// getRecentConversationsAcrossDialogs 跨Dialog获取最近的conversations
func getRecentConversationsAcrossDialogs(sessionID int64, limit int) ([]models.ConversationModel, error) {
	var conversations []models.ConversationModel
	err := global.DB.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&conversations).Error
	return conversations, err
}

// traceParentConversations 从指定对话节点往上追溯父节点
func traceParentConversations(dialogID int64, maxLayers int) ([]models.ConversationModel, error) {
	// 保持向后兼容：从dialog的最新conversation开始追溯
	var latestConv models.ConversationModel
	err := global.DB.Where("dialog_id = ?", dialogID).
		Order("created_at DESC").
		First(&latestConv).Error
	if err != nil {
		return nil, err
	}

	return traceParentConversationsFromConversation(latestConv.ID, maxLayers)
}

// traceParentConversationsFromConversation 从指定conversation开始往上追溯父节点
// 这是核心实现，能正确处理分叉路径
func traceParentConversationsFromConversation(conversationID int64, maxLayers int) ([]models.ConversationModel, error) {
	var conversations []models.ConversationModel
	currentConversationID := &conversationID

	for i := 0; i < maxLayers && currentConversationID != nil; i++ {
		// 获取当前conversation
		var conv models.ConversationModel
		err := global.DB.First(&conv, *currentConversationID).Error
		if err != nil {
			if err.Error() == "record not found" {
				break
			}
			return nil, fmt.Errorf("获取conversation失败: %v", err)
		}

		conversations = append(conversations, conv)

		// 查找父conversation
		parentConversation, err := findParentConversation(conv)
		if err != nil {
			// 如果找不到父conversation，说明已经到达根节点
			logrus.Debugf("已找到达根节点，或其他错误: %v", err)
			break
		}
		logrus.Debugf("找到父对话，id: %d", parentConversation.ID)
		currentConversationID = &parentConversation.ID
	}

	return conversations, nil
}

// findParentConversation 找到指定conversation的父conversation
func findParentConversation(conv models.ConversationModel) (*models.ConversationModel, error) {
	// 首先在同一dialog内查找前一个conversation
	var prevConversation models.ConversationModel
	err := global.DB.Where("dialog_id = ? AND created_at < ?", conv.DialogID, conv.CreatedAt).
		Order("created_at DESC").
		First(&prevConversation).Error

	if err == nil {
		// 找到同一dialog内的前一个conversation
		return &prevConversation, nil
	}

	// 如果同一dialog内没有更早的conversation，则查找跨dialog的父conversation
	// 获取当前conversation所在的dialog
	var currentDialog models.DialogModel
	err = global.DB.First(&currentDialog, conv.DialogID).Error
	if err != nil {
		return nil, fmt.Errorf("获取dialog失败: %v", err)
	}

	// 如果当前dialog没有父dialog，说明已经是根节点
	if currentDialog.ParentID == nil {
		return nil, fmt.Errorf("已到达根节点")
	}

	// 如果当前dialog有分叉点信息，直接使用分叉点conversation
	if currentDialog.BranchFromConversationID != nil {
		var parentConversation models.ConversationModel
		err = global.DB.First(&parentConversation, *currentDialog.BranchFromConversationID).Error
		if err != nil {
			return nil, fmt.Errorf("查询分叉点conversation失败: %v", err)
		}
		return &parentConversation, nil
	}

	// 兼容旧数据：如果没有分叉点信息，使用原来的逻辑
	// 找到父dialog中的分叉点conversation
	// 在分叉场景下，我们需要找到分叉发生时的那个conversation
	// 策略：查找父dialog中在当前dialog创建之前就存在的最新conversation
	var parentConversation models.ConversationModel

	// 首先尝试找到父dialog中在当前dialog创建时间之前的最新conversation
	err = global.DB.Where("dialog_id = ? AND created_at < ?",
		*currentDialog.ParentID, currentDialog.CreatedAt).
		Order("created_at DESC").
		First(&parentConversation).Error

	if err != nil {
		if err.Error() == "record not found" {
			// 如果找不到时间在前的conversation，说明当前dialog是从父dialog的最新conversation分叉的
			// 这种情况下，直接找父dialog的最新conversation
			err = global.DB.Where("dialog_id = ?", *currentDialog.ParentID).
				Order("created_at DESC").
				First(&parentConversation).Error
			if err != nil {
				return nil, fmt.Errorf("找不到父conversation: %v", err)
			}
		} else {
			return nil, fmt.Errorf("查询父conversation失败: %v", err)
		}
	}

	return &parentConversation, nil
}

// buildLongTermContext 构建长期记忆上下文（向量检索）
func buildLongTermContext(sessionID int64, currentQuestion string) (string, error) {
	if !global.Config.Vector.Enable {
		return "", nil
	}

	// 1. 对当前问题进行向量化
	questionVector, err := embedding_service.GetEmbedding(currentQuestion)
	if err != nil {
		return "", fmt.Errorf("问题向量化失败: %v", err)
	}

	// 2. 在向量数据库中检索相似的历史对话
	filter := map[string]interface{}{
		"session_id": sessionID,
	}

	results, err := vector_service.VectorServiceInstance.Search(
		questionVector,
		global.Config.Vector.TopK,
		filter,
	)
	if err != nil {
		return "", fmt.Errorf("向量检索失败: %v", err)
	}

	if len(results) == 0 {
		return "", nil
	}

	// 3. 构建长期记忆上下文
	var contextLines []string
	for _, result := range results {
		// 从向量数据库元数据中获取 conversation_id
		conversationID := result.ID // result.ID 已经是 uint64 类型的 conversation_id

		// 从主数据库中查询对应的 ConversationModel
		var conversation models.ConversationModel
		err := global.DB.First(&conversation, conversationID).Error
		if err != nil {
			// 如果找不到对应的 conversation，记录错误并跳过
			fmt.Printf("Warning: Conversation with ID %d not found in DB: %v\n", conversationID, err)
			continue
		}

		// 使用从数据库中获取的 prompt 和 summary
		contextLines = append(contextLines, fmt.Sprintf("历史相关问题: %s", conversation.Prompt))
		contextLines = append(contextLines, fmt.Sprintf("回答要点: %s", conversation.Summary))
		contextLines = append(contextLines, "---")
	}

	if len(contextLines) == 0 {
		return "", nil
	}

	return strings.Join(contextLines, "\n"), nil
}

// StoreConversationVector 将对话存储到向量数据库
func StoreConversationVector(conversationID int64, prompt, answer, summary string) error {
	if !global.Config.Vector.Enable {
		return nil
	}

	// 1. 获取对话的详细信息
	var conversation models.ConversationModel
	err := global.DB.Preload("SessionModel").First(&conversation, conversationID).Error
	if err != nil {
		return fmt.Errorf("获取对话信息失败: %v", err)
	}

	// 2. 对问题和回答进行向量化（这里选择对问题进行向量化，因为检索时主要是基于问题的相似性）
	questionVector, err := embedding_service.GetEmbedding(prompt)
	if err != nil {
		return fmt.Errorf("问题向量化失败: %v", err)
	}

	// 3. 准备元数据
	metadata := map[string]interface{}{
		"conversation_id": conversationID,
		"session_id":      conversation.SessionID,
		"dialog_id":       conversation.DialogID,
	}

	// 4. 存储到向量数据库，ID使用uint64类型
	err = vector_service.VectorServiceInstance.Store(uint64(conversationID), questionVector, metadata)
	if err != nil {
		return fmt.Errorf("向量存储失败: %v", err)
	}

	return nil
}

// DeleteConversationVector 从向量数据库删除对话
func DeleteConversationVector(conversationID int64) error {
	if !global.Config.Vector.Enable {
		return nil
	}

	return vector_service.VectorServiceInstance.Delete(uint64(conversationID))
}

// DeleteSessionVectors 删除整个会话的所有向量
func DeleteSessionVectors(sessionID int64) error {
	if !global.Config.Vector.Enable {
		return nil
	}

	// 获取会话下的所有对话ID
	var conversations []models.ConversationModel
	err := global.DB.Where("session_id = ?", sessionID).Find(&conversations).Error
	if err != nil {
		return err
	}

	// 逐一删除向量
	for _, conv := range conversations {
		err := vector_service.VectorServiceInstance.Delete(uint64(conv.ID))
		if err != nil {
			// 记录错误但继续删除其他向量
			fmt.Printf("删除向量失败 %d: %v\n", conv.ID, err)
		}
	}

	return nil
}

// FindParentConversation 根据ParentDialogID找到要作为父节点的conversation
func FindParentConversation(parentDialogID int64) (*models.ConversationModel, error) {
	var conversation models.ConversationModel
	err := global.DB.Where("dialog_id = ?", parentDialogID).
		Order("created_at DESC").
		First(&conversation).Error
	if err != nil {
		return nil, fmt.Errorf("找不到父节点conversation: %v", err)
	}
	return &conversation, nil
}

// CheckIfBranching 检测是否需要分叉
// 如果选择的父conversation不是当前dialog的最新conversation，则需要分叉
func CheckIfBranching(sessionID int64, parentDialogID *int64) (bool, *models.ConversationModel, error) {
	if parentDialogID == nil {
		// 没有指定父节点，不需要分叉
		return false, nil, nil
	}

	// 找到指定的父conversation
	parentConv, err := FindParentConversation(*parentDialogID)
	if err != nil {
		return false, nil, err
	}

	// 找到当前dialog中最新的conversation
	var latestConv models.ConversationModel
	err = global.DB.Where("dialog_id = ?", *parentDialogID).
		Order("created_at DESC").
		First(&latestConv).Error
	if err != nil {
		return false, nil, fmt.Errorf("获取最新conversation失败: %v", err)
	}

	// 如果选择的父conversation不是最新的，则需要分叉
	needsBranching := parentConv.ID != latestConv.ID
	return needsBranching, parentConv, nil
}

// BuildDialogContextFromConversation 根据conversation ID构建对话上下文
// 这个函数能正确处理分叉场景下的上下文追溯，返回JSON格式
func BuildDialogContextFromConversation(sessionID int64, parentConversationID *int64, currentQuestion string) (string, error) {
	contextData := ContextData{
		Recent:  []QAPair{},
		History: []QAPair{},
		Current: currentQuestion,
	}

	// 1. 构建短期记忆上下文（从指定conversation往上追溯）
	recentConversations, err := getRecentConversationsFromConversation(sessionID, parentConversationID)
	if err != nil {
		return "", fmt.Errorf("构建短期上下文失败: %v", err)
	}

	// 将最近对话转换为QAPair格式（按时间正序排列）
	for i := len(recentConversations) - 1; i >= 0; i-- {
		conv := recentConversations[i]
		contextData.Recent = append(contextData.Recent, QAPair{
			Q: truncateText(conv.Prompt, "提问中间部分"),
			A: truncateText(conv.Answer, "答案中间部分"), // 使用完整回答
		})
	}

	// 2. 构建长期记忆上下文（向量检索相关历史）
	historyConversations, err := getLongTermContextConversations(sessionID, currentQuestion)
	if err != nil {
		// 长期记忆检索失败不应该影响整个对话流程，只记录错误
		fmt.Printf("长期记忆检索失败: %v\n", err)
	} else {
		contextData.History = historyConversations
	}

	// 3. 序列化为JSON
	jsonData, err := json.Marshal(contextData)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	logrus.Debug("\n" + strings.Repeat("=", 30) + "上下文拼接开始" + strings.Repeat("=", 30) + "\n")
	logrus.Debugf("本次Recent: %v+\n", contextData.Recent)
	logrus.Debugf("本次History: %v+\n", contextData.History)
	logrus.Debugf("本次Current: %s\n", contextData.Current)
	logrus.Debug(strings.Repeat("=", 30) + "上下文拼接结束" + strings.Repeat("=", 30) + "\n")

	return string(jsonData), nil
}

func truncateText(text, note string) string {
	if len(text) <= maxLength {
		return text
	}
	head := text[:headTailLength]
	tail := text[len(text)-headTailLength:]
	return fmt.Sprintf("%s...(%s已省略)...%s", head, note, tail)
}

// getRecentConversationsFromConversation 从指定conversation获取最近的对话记录
func getRecentConversationsFromConversation(sessionID int64, parentConversationID *int64) ([]models.ConversationModel, error) {
	contextLayers := global.Config.Ai.ContextLayers
	if contextLayers <= 0 {
		return []models.ConversationModel{}, nil
	}

	if parentConversationID == nil {
		// 如果没有指定父conversation，获取会话中最新的几轮对话（跨Dialog）
		return getRecentConversationsAcrossDialogs(sessionID, contextLayers)
	} else {
		// 从指定的conversation开始往上追溯
		return traceParentConversationsFromConversation(*parentConversationID, contextLayers)
	}
}

// getLongTermContextConversations 获取长期记忆相关对话
func getLongTermContextConversations(sessionID int64, currentQuestion string) ([]QAPair, error) {
	if !global.Config.Vector.Enable {
		return []QAPair{}, nil
	}

	// 1. 对当前问题进行向量化
	questionVector, err := embedding_service.GetEmbedding(currentQuestion)
	if err != nil {
		return nil, fmt.Errorf("问题向量化失败: %v", err)
	}

	// 2. 在向量数据库中检索相似的历史对话
	filter := map[string]interface{}{
		"session_id": sessionID,
	}

	results, err := vector_service.VectorServiceInstance.Search(
		questionVector,
		global.Config.Vector.TopK,
		filter,
	)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %v", err)
	}

	// 3. 构建历史对话QAPair列表
	var historyPairs []QAPair
	for _, result := range results {
		// 从向量数据库元数据中获取 conversation_id (result.ID 已经是 uint64 类型的 conversation_id)
		conversationID := result.ID

		// 从主数据库中查询对应的 ConversationModel
		var conversation models.ConversationModel
		err := global.DB.First(&conversation, conversationID).Error
		if err != nil {
			// 如果找不到对应的 conversation，记录错误并跳过
			fmt.Printf("Warning: Conversation with ID %d not found in DB for long-term context: %v\n", conversationID, err)
			continue
		}

		historyPairs = append(historyPairs, QAPair{
			Q: conversation.Prompt,
			A: conversation.Summary,
		})
	}

	return historyPairs, nil
}

// CheckIfBranchingByConversation 根据conversation ID检测是否需要分叉
func CheckIfBranchingByConversation(parentConversationID int64) (bool, error) {
	// 获取父conversation
	var parentConv models.ConversationModel
	if err := global.DB.First(&parentConv, parentConversationID).Error; err != nil {
		return false, fmt.Errorf("找不到父conversation: %v", err)
	}

	// 找到同一dialog中最新的conversation
	var latestConv models.ConversationModel
	err := global.DB.Where("dialog_id = ?", parentConv.DialogID).
		Order("created_at DESC").
		First(&latestConv).Error
	if err != nil {
		return false, fmt.Errorf("获取最新conversation失败: %v", err)
	}

	// 如果指定的父conversation不是最新的，则需要分叉
	return parentConv.ID != latestConv.ID, nil
}

// CreateBranchingDialogs 创建分叉时的新dialogs
// 返回: 新对话的dialogID, 被分叉出去的conversations的新dialogID, error
func CreateBranchingDialogs(sessionID int64, parentConversationID int64, parentDialogID int64) (int64, int64, error) {
	tx := global.DB.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if !committed {
			tx.Rollback()
		}
	}()

	// 1. 创建新 dialog，用于用户输入新分支
	newDialog := models.DialogModel{
		SessionID:                sessionID,
		ParentID:                 &parentDialogID,
		BranchFromConversationID: &parentConversationID,
	}
	if err := tx.Create(&newDialog).Error; err != nil {
		return 0, 0, fmt.Errorf("创建新 dialog 失败: %v", err)
	}

	logrus.Debugf("\n创建新 dialog，用于用户输入新分支 %v+\n", newDialog)

	// 2. 创建另一个 dialog，用于接收被分出的历史对话（分叉点之后）
	branchedDialog := models.DialogModel{
		SessionID:                sessionID,
		ParentID:                 &parentDialogID,
		BranchFromConversationID: &parentConversationID,
	}
	if err := tx.Create(&branchedDialog).Error; err != nil {
		return 0, 0, fmt.Errorf("创建分支 dialog 失败: %v", err)
	}

	logrus.Debugf("\n创建另一个 dialog，用于接收被分出的历史对话（分叉点之后） %v+\n", branchedDialog)

	// 3. 将原来 parentDialogID 的子 dialog 的 parent_id 改为 branchedDialog.ID
	if err := tx.Model(&models.DialogModel{}).
		Where("parent_id = ? AND id NOT IN ?", parentDialogID, []int64{newDialog.ID, branchedDialog.ID}).
		Update("parent_id", branchedDialog.ID).Error; err != nil {
		return 0, 0, fmt.Errorf("更新子 dialog 的父 ID 失败: %v", err)
	}

	logrus.Debugf("将原本父 dialog %d 的 dialog，修改其父节点 id 为 %d\n", parentDialogID, branchedDialog.ID)

	// 4. 检查分叉点 conversation 是否存在（可选：防止 parentID 非法）
	var parentConv models.ConversationModel
	if err := tx.First(&parentConv, parentConversationID).Error; err != nil {
		return 0, 0, fmt.Errorf("分叉点 conversation 不存在: %v", err)
	}

	// 5. 移动分叉点之后的对话记录（使用 ID 比时间更可靠）
	if err := tx.Model(&models.ConversationModel{}).
		Where("dialog_id = ? AND id > ?", parentDialogID, parentConversationID).
		Update("dialog_id", branchedDialog.ID).Error; err != nil {
		return 0, 0, fmt.Errorf("移动 conversation 到分支失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, 0, fmt.Errorf("提交事务失败: %v", err)
	}
	committed = true

	return newDialog.ID, branchedDialog.ID, nil
}
