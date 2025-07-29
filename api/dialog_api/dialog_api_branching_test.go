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

// setupTestConfig 设置测试配置
func setupTestConfig() {
	if global.Config == nil {
		global.Config = &conf.Config{
			Ai: conf.Ai{
				ContextLayers: 10,
			},
			Vector: conf.Vector{
				Enable: false,
			},
		}
	}
}

// setupAPITestDB 设置API测试数据库
func setupAPITestDB(t *testing.T) *gorm.DB {
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

	return db
}

// setupAPITestData 创建API测试数据
func setupAPITestData(t *testing.T, db *gorm.DB) (int64, int64, []int64) {
	// 创建类别
	category := models.CategoryModel{
		Name: "测试类别",
	}
	category.ID = 1
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	db.Create(&category)

	// 创建会话
	session := models.SessionModel{
		Tittle:     "测试会话",
		Summary:    "测试摘要",
		CategoryID: category.ID,
	}
	session.ID = 1
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	db.Create(&session)

	// 创建对话树
	dialog := models.DialogModel{
		SessionID: session.ID,
		ParentID:  nil,
	}
	dialog.ID = 1
	dialog.CreatedAt = time.Now()
	dialog.UpdatedAt = time.Now()
	db.Create(&dialog)

	// 更新session的root_dialog_id
	session.RootDialogID = &dialog.ID
	db.Save(&session)

	// 创建3个conversations: c1->c2->c3
	var conversationIDs []int64
	for i := 0; i < 3; i++ {
		conv := models.ConversationModel{
			Prompt:    fmt.Sprintf("问题%d", i+1),
			Answer:    fmt.Sprintf("回答%d", i+1),
			SessionID: session.ID,
			DialogID:  dialog.ID,
			Title:     fmt.Sprintf("标题%d", i+1),
			Summary:   fmt.Sprintf("摘要%d", i+1),
			IsStarred: false,
			Comment:   "",
		}
		conv.ID = int64(i + 1)
		conv.CreatedAt = time.Now().Add(time.Duration(i) * time.Minute)
		conv.UpdatedAt = conv.CreatedAt
		db.Create(&conv)
		conversationIDs = append(conversationIDs, conv.ID)
	}

	return session.ID, dialog.ID, conversationIDs
}

// TestAPIBranchingScenario 测试通过API进行分叉
func TestAPIBranchingScenario(t *testing.T) {
	// 设置测试环境
	setupTestConfig() // 使用内部配置
	db := setupAPITestDB(t)
	global.DB = db

	// 创建测试数据
	sessionID, _, conversationIDs := setupAPITestData(t, db)

	// 创建gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 注册路由
	dialogAPI := DialogApi{}
	router.POST("/dialog/new-chat-sync", dialogAPI.NewChatSync)

	t.Logf("初始状态:")
	t.Logf("  sessionID: %d", sessionID)
	t.Logf("  conversationIDs: %v", conversationIDs)

	// 验证初始状态
	var initialConvs []models.ConversationModel
	db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&initialConvs)
	t.Logf("初始conversations数量: %d", len(initialConvs))
	for i, conv := range initialConvs {
		t.Logf("  [%d] %s (ID: %d, DialogID: %d)", i, conv.Prompt, conv.ID, conv.DialogID)
	}

	// 准备分叉请求 - 选择c2(ID=2)作为父节点
	parentConversationID := conversationIDs[1] // c2
	requestData := NewChatReq{
		Content:              "新问题4",
		SessionID:            sessionID,
		ParentConversationID: &parentConversationID,
	}

	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/dialog/new-chat-sync", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("API响应状态: %d", w.Code)
	t.Logf("API响应内容: %s", w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("API调用失败，状态码: %d, 响应: %s", w.Code, w.Body.String())
	}

	// 解析响应
	var response struct {
		Code int          `json:"code"`
		Data ChatResponse `json:"data"`
		Msg  string       `json:"msg"`
	}
	
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if response.Code != 0 { // 0是成功代码
		t.Fatalf("API返回错误: %s", response.Msg)
	}

	t.Logf("新对话创建成功:")
	t.Logf("  DialogID: %d", response.Data.DialogID)
	t.Logf("  ConversationID: %d", response.Data.ConversationID)

	// 验证分叉后的数据结构
	t.Run("验证分叉后的数据结构", func(t *testing.T) {
		// 查看所有dialogs
		var allDialogs []models.DialogModel
		db.Find(&allDialogs)
		
		t.Logf("分叉后的所有dialogs:")
		for _, dialog := range allDialogs {
			t.Logf("  Dialog ID: %d, SessionID: %d, ParentID: %v, BranchFrom: %v", 
				dialog.ID, dialog.SessionID, dialog.ParentID, dialog.BranchFromConversationID)
		}

		// 查看所有conversations的分布
		var allConversations []models.ConversationModel
		db.Order("dialog_id ASC, created_at ASC").Find(&allConversations)

		t.Logf("分叉后所有conversations的分布:")
		currentDialogID := int64(-1)
		for _, conv := range allConversations {
			if conv.DialogID != currentDialogID {
				t.Logf("  Dialog %d:", conv.DialogID)
				currentDialogID = conv.DialogID
			}
			t.Logf("    - %s (ID: %d, Created: %s)", conv.Prompt, conv.ID, conv.CreatedAt.Format("15:04:05"))
		}

		// 验证期望的结构
		// 应该有3个dialogs: 原始(1), 新对话(2), 分叉出来的(3)
		if len(allDialogs) != 3 {
			t.Errorf("期望有3个dialogs，实际有%d个", len(allDialogs))
		}

		// 验证原始dialog只有c1和c2
		var originalDialogConvs []models.ConversationModel
		db.Where("dialog_id = ?", 1).Order("created_at ASC").Find(&originalDialogConvs)
		
		if len(originalDialogConvs) != 2 {
			t.Errorf("原始dialog应该有2个conversations，实际有%d个", len(originalDialogConvs))
		}

		// 验证新对话dialog有新的conversation
		var newDialogConvs []models.ConversationModel
		db.Where("dialog_id = ?", response.Data.DialogID).Find(&newDialogConvs)
		
		if len(newDialogConvs) != 1 {
			t.Errorf("新dialog应该有1个conversation，实际有%d个", len(newDialogConvs))
		}

		// 验证分叉dialog有c3
		var branchedDialogConvs []models.ConversationModel
		db.Where("dialog_id != ? AND dialog_id != ?", 1, response.Data.DialogID).Find(&branchedDialogConvs)
		
		if len(branchedDialogConvs) != 1 {
			t.Errorf("分叉dialog应该有1个conversation，实际有%d个", len(branchedDialogConvs))
		} else {
			if branchedDialogConvs[0].Prompt != "问题3" {
				t.Errorf("分叉dialog应该包含问题3，实际包含%s", branchedDialogConvs[0].Prompt)
			}
		}
	})
}