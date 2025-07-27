package dialog_service

import (
	"dialogTree/conf"
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestConfig 设置测试配置
func setupTestConfig() {
	if global.Config == nil {
		// 尝试读取配置文件，如果失败则使用默认配置
		defer func() {
			if r := recover(); r != nil {
				// 配置文件读取失败，使用默认配置
				global.Config = &conf.Config{
					Ai: conf.Ai{
						ContextLayers: 3,
					},
					Vector: conf.Vector{
						Enable: false, // 测试中不启用向量数据库
					},
				}
			}
		}()
		
		global.Config = core.ReadConf(true)
	}
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
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

// createTestData 创建测试数据
func createTestData(t *testing.T, db *gorm.DB) (int64, int64, []int64) {
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

	// 创建多个conversations（模拟顺序对话）
	conversationIDs := make([]int64, 5)
	for i := 0; i < 5; i++ {
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

// TestCheckIfBranchingByConversation 测试基于conversation的分叉检测
func TestCheckIfBranchingByConversation(t *testing.T) {
	db := setupTestDB(t)
	global.DB = db

	sessionID, dialogID, conversationIDs := createTestData(t, db)
	_ = sessionID
	_ = dialogID

	// 测试用例1：选择最新的conversation（不应该分叉）
	t.Run("最新conversation不分叉", func(t *testing.T) {
		latestConvID := conversationIDs[len(conversationIDs)-1] // 最后一个
		needsBranching, err := CheckIfBranchingByConversation(latestConvID)
		if err != nil {
			t.Errorf("检查分叉失败: %v", err)
		}
		if needsBranching {
			t.Error("最新conversation不应该需要分叉")
		}
	})

	// 测试用例2：选择历史conversation（应该分叉）
	t.Run("历史conversation需要分叉", func(t *testing.T) {
		historicalConvID := conversationIDs[2] // 中间的一个
		needsBranching, err := CheckIfBranchingByConversation(historicalConvID)
		if err != nil {
			t.Errorf("检查分叉失败: %v", err)
		}
		if !needsBranching {
			t.Error("历史conversation应该需要分叉")
		}
	})

	// 测试用例3：不存在的conversation
	t.Run("不存在conversation报错", func(t *testing.T) {
		_, err := CheckIfBranchingByConversation(999)
		if err == nil {
			t.Error("不存在的conversation应该报错")
		}
	})
}

// TestCreateBranchingDialogs 测试分叉创建逻辑
func TestCreateBranchingDialogs(t *testing.T) {
	db := setupTestDB(t)
	global.DB = db

	sessionID, dialogID, conversationIDs := createTestData(t, db)

	// 选择中间的conversation作为分叉点
	branchPointConvID := conversationIDs[2] // 第3个conversation

	// 执行分叉
	newDialogID, branchedDialogID, err := CreateBranchingDialogs(sessionID, branchPointConvID, dialogID)
	if err != nil {
		t.Fatalf("创建分叉失败: %v", err)
	}

	// 验证结果
	t.Run("验证新dialog创建", func(t *testing.T) {
		var newDialog models.DialogModel
		err := db.First(&newDialog, newDialogID).Error
		if err != nil {
			t.Errorf("新dialog不存在: %v", err)
		}
		if newDialog.SessionID != sessionID {
			t.Errorf("新dialog的sessionID不正确")
		}
		if newDialog.ParentID == nil || *newDialog.ParentID != dialogID {
			t.Errorf("新dialog的parentID不正确")
		}
	})

	t.Run("验证分叉dialog创建", func(t *testing.T) {
		var branchedDialog models.DialogModel
		err := db.First(&branchedDialog, branchedDialogID).Error
		if err != nil {
			t.Errorf("分叉dialog不存在: %v", err)
		}
		if branchedDialog.SessionID != sessionID {
			t.Errorf("分叉dialog的sessionID不正确")
		}
		if branchedDialog.ParentID == nil || *branchedDialog.ParentID != dialogID {
			t.Errorf("分叉dialog的parentID不正确")
		}
	})

	t.Run("验证conversations迁移", func(t *testing.T) {
		// 检查分叉点之后的conversations是否被移动到新的dialog
		var movedConversations []models.ConversationModel
		db.Where("dialog_id = ?", branchedDialogID).Find(&movedConversations)

		// 应该有两个conversations被移动（第4和第5个）
		if len(movedConversations) != 2 {
			t.Errorf("移动的conversations数量不正确，期望2个，实际%d个", len(movedConversations))
		}

		// 检查原始dialog中剩余的conversations
		var remainingConversations []models.ConversationModel
		db.Where("dialog_id = ?", dialogID).Find(&remainingConversations)

		// 应该剩余3个conversations（第1、第2、第3个）
		if len(remainingConversations) != 3 {
			t.Errorf("剩余的conversations数量不正确，期望3个，实际%d个", len(remainingConversations))
		}
	})
}

// TestCrossDialogContext 测试跨Dialog上下文查询
func TestCrossDialogContext(t *testing.T) {
	db := setupTestDB(t)
	global.DB = db

	sessionID, _, _ := createTestData(t, db)

	// 测试跨Dialog获取最近的conversations
	t.Run("跨Dialog获取conversations", func(t *testing.T) {
		conversations, err := getRecentConversationsAcrossDialogs(sessionID, 3)
		if err != nil {
			t.Errorf("获取conversations失败: %v", err)
		}
		if len(conversations) != 3 {
			t.Errorf("获取的conversations数量不正确，期望3个，实际%d个", len(conversations))
		}

		// 验证按时间降序排列
		for i := 1; i < len(conversations); i++ {
			if conversations[i-1].CreatedAt.Before(conversations[i].CreatedAt) {
				t.Error("获取的conversations没有按时间降序排列")
			}
		}
	})
}

// TestBuildDialogContextFromConversation 测试从指定conversation构建上下文
func TestBuildDialogContextFromConversation(t *testing.T) {
	setupTestConfig()
	db := setupTestDB(t)
	global.DB = db

	sessionID, _, conversationIDs := createTestData(t, db)

	t.Run("从指定conversation构建上下文", func(t *testing.T) {
		parentConvID := conversationIDs[2]
		context, err := BuildDialogContextFromConversation(sessionID, &parentConvID, "新问题")
		if err != nil {
			t.Errorf("构建上下文失败: %v", err)
		}
		if context == "" {
			t.Error("上下文不应该为空")
		}
		t.Logf("构建的上下文: %s", context)
	})

	t.Run("未指定conversation构建上下文", func(t *testing.T) {
		context, err := BuildDialogContextFromConversation(sessionID, nil, "新问题")
		if err != nil {
			t.Errorf("构建上下文失败: %v", err)
		}
		// 可以为空，也可以不为空，取决于数据
		t.Logf("构建的上下文: %s", context)
	})
}

// TestCompleteWorkflow 测试完整的分叉工作流程
func TestCompleteWorkflow(t *testing.T) {
	setupTestConfig()
	db := setupTestDB(t)
	global.DB = db

	sessionID, dialogID, conversationIDs := createTestData(t, db)

	t.Run("完整分叉流程", func(t *testing.T) {
		// 步骤1：选择历史conversation作为分叉点
		branchPointConvID := conversationIDs[1] // 第2个conversation

		// 步骤2：检查是否需要分叉
		needsBranching, err := CheckIfBranchingByConversation(branchPointConvID)
		if err != nil {
			t.Fatalf("检查分叉失败: %v", err)
		}
		if !needsBranching {
			t.Fatal("应该需要分叉")
		}

		// 步骤3：执行分叉
		newDialogID, branchedDialogID, err := CreateBranchingDialogs(sessionID, branchPointConvID, dialogID)
		if err != nil {
			t.Fatalf("创建分叉失败: %v", err)
		}

		// 步骤4：验证分叉后的数据结构
		// 原始dialog应该只有前2个conversations
		var originalConversations []models.ConversationModel
		db.Where("dialog_id = ?", dialogID).Order("created_at ASC").Find(&originalConversations)
		if len(originalConversations) != 2 {
			t.Errorf("原始dialog应该有2个conversations，实际%d个", len(originalConversations))
		}

		// 分叉dialog应该有后3个conversations
		var branchedConversations []models.ConversationModel
		db.Where("dialog_id = ?", branchedDialogID).Order("created_at ASC").Find(&branchedConversations)
		if len(branchedConversations) != 3 {
			t.Errorf("分叉dialog应该有3个conversations，实际%d个", len(branchedConversations))
		}

		// 新dialog暂时应该为空（用于存放新对话）
		var newConversations []models.ConversationModel
		db.Where("dialog_id = ?", newDialogID).Find(&newConversations)
		if len(newConversations) != 0 {
			t.Errorf("新dialog应该为空，实际有%d个conversations", len(newConversations))
		}

		// 步骤5：验证上下文构建
		context, err := BuildDialogContextFromConversation(sessionID, &branchPointConvID, "新问题")
		if err != nil {
			t.Errorf("构建上下文失败: %v", err)
		}
		t.Logf("分叉后的上下文: %s", context)
	})
}