package dialog_api

import (
	"bytes"
	"dialogTree/conf"
	"dialogTree/global"
	"dialogTree/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestEnvironment 设置测试环境
func setupTestEnvironment(t *testing.T) (*gorm.DB, *gin.Engine) {
	// 设置测试配置
	global.Config = &conf.Config{
		Ai: conf.Ai{
			ContextLayers: 3,
			ChatAnywhere: conf.ChatAnywhere{
				SecretKey: "", // 空密钥用于测试
				Model:     "test-model",
			},
		},
		Vector: conf.Vector{
			Enable: false, // 测试中不启用向量数据库
		},
	}

	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.SessionModel{},
		&models.DialogModel{},
		&models.ConversationModel{},
		&models.CategoryModel{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	global.DB = db

	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册路由
	dialogApi := DialogApi{}
	router.POST("/api/dialog/chat/sync", dialogApi.NewChatSync)

	return db, router
}

// createTestSessionAndDialog 创建测试会话和对话
func createTestSessionAndDialog(t *testing.T, db *gorm.DB) (int64, int64, []int64) {
	// 创建类别
	category := models.CategoryModel{
		Model: models.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:  "测试类别",
	}
	db.Create(&category)

	// 创建会话
	session := models.SessionModel{
		Model:      models.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Tittle:     "测试会话",
		Summary:    "测试摘要",
		CategoryID: category.ID,
	}
	db.Create(&session)

	// 创建对话树
	dialog := models.DialogModel{
		Model:     models.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		SessionID: session.ID,
		ParentID:  nil,
	}
	db.Create(&dialog)

	// 更新session的root_dialog_id
	session.RootDialogID = &dialog.ID
	db.Save(&session)

	// 创建一些初始conversations
	conversationIDs := make([]int64, 3)
	for i := 0; i < 3; i++ {
		conv := models.ConversationModel{
			Model:     models.Model{ID: int64(i + 1), CreatedAt: time.Now().Add(time.Duration(i) * time.Minute), UpdatedAt: time.Now()},
			Prompt:    fmt.Sprintf("问题%d", i+1),
			Answer:    fmt.Sprintf("回答%d", i+1),
			SessionID: session.ID,
			DialogID:  dialog.ID,
			Title:     fmt.Sprintf("标题%d", i+1),
			Summary:   fmt.Sprintf("摘要%d", i+1),
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv)
		conversationIDs[i] = conv.ID
	}

	return session.ID, dialog.ID, conversationIDs
}

// TestNewChatSync_NormalFlow 测试正常对话流程（不分叉）
func TestNewChatSync_NormalFlow(t *testing.T) {
	db, router := setupTestEnvironment(t)
	sessionID, _, conversationIDs := createTestSessionAndDialog(t, db)

	// 准备请求：从最新conversation继续对话（不应该分叉）
	reqBody := NewChatReq{
		Content:              "新问题",
		SessionID:            sessionID,
		ParentConversationID: &conversationIDs[2], // 最新的conversation
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/dialog/chat/sync", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
	}

	// 验证不应该创建新的dialog（因为没有分叉）
	var dialogs []models.DialogModel
	db.Find(&dialogs)
	if len(dialogs) != 1 { // 应该还是只有原来的1个dialog
		t.Errorf("不应该创建新dialog，实际dialogs数量：%d", len(dialogs))
	}

	// 验证应该创建新的conversation
	var conversations []models.ConversationModel
	db.Where("dialog_id = ?", dialogs[0].ID).Find(&conversations)
	if len(conversations) != 4 { // 原来3个 + 新的1个
		t.Errorf("应该有4个conversations，实际：%d", len(conversations))
	}
}

// TestNewChatSync_BranchingFlow 测试分叉对话流程
func TestNewChatSync_BranchingFlow(t *testing.T) {
	db, router := setupTestEnvironment(t)
	sessionID, _, conversationIDs := createTestSessionAndDialog(t, db)

	// 准备请求：从历史conversation分叉对话
	reqBody := NewChatReq{
		Content:              "分叉问题",
		SessionID:            sessionID,
		ParentConversationID: &conversationIDs[1], // 第2个conversation（不是最新的）
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/dialog/chat/sync", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
	}

	// 验证应该创建新的dialogs（分叉）
	var dialogs []models.DialogModel
	db.Find(&dialogs)
	if len(dialogs) != 3 { // 原来1个 + 新的2个（新对话dialog + 被分叉出去的dialog）
		t.Errorf("应该创建2个新dialogs，实际dialogs数量：%d", len(dialogs))
	}

	// 验证原始dialog中只剩下前2个conversations
	originalDialog := dialogs[0] // 第一个是原始dialog
	var originalConversations []models.ConversationModel
	db.Where("dialog_id = ?", originalDialog.ID).Find(&originalConversations)
	if len(originalConversations) != 2 {
		t.Errorf("原始dialog应该有2个conversations，实际：%d", len(originalConversations))
	}

	// 验证新的对话dialog中有新创建的conversation
	var newDialogs []models.DialogModel
	db.Where("id != ?", originalDialog.ID).Find(&newDialogs)
	
	var newConversationExists bool
	for _, dialog := range newDialogs {
		var conversations []models.ConversationModel
		db.Where("dialog_id = ?", dialog.ID).Find(&conversations)
		for _, conv := range conversations {
			if conv.Prompt == "分叉问题" {
				newConversationExists = true
				break
			}
		}
		if newConversationExists {
			break
		}
	}

	if !newConversationExists {
		t.Error("新对话应该被创建在新的dialog中")
	}
}

// TestNewChatSync_WithoutParent 测试没有指定父conversation的情况
func TestNewChatSync_WithoutParent(t *testing.T) {
	db, router := setupTestEnvironment(t)
	sessionID, _, _ := createTestSessionAndDialog(t, db)

	// 准备请求：不指定ParentConversationID
	reqBody := NewChatReq{
		Content:              "新会话问题",
		SessionID:            sessionID,
		ParentConversationID: nil, // 不指定父节点
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/dialog/chat/sync", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
	}

	// 验证应该创建新的根dialog
	var dialogs []models.DialogModel
	db.Where("parent_id IS NULL").Find(&dialogs)
	if len(dialogs) != 2 { // 原来1个根dialog + 新的1个根dialog
		t.Errorf("应该有2个根dialogs，实际：%d", len(dialogs))
	}
}