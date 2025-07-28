package dialog_service

import (
	"dialogTree/conf"
	"dialogTree/core"
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"strings"
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

// TestBranchPointTracing 测试在寻找父conversation过程中遇到branch point的情况
func TestBranchPointTracing(t *testing.T) {
	setupTestConfig()
	db := setupTestDB(t)
	global.DB = db

	sessionID, rootDialogID, _ := createTestData(t, db)

	// 创建更复杂的分叉场景
	// Dialog结构:
	// rootDialog (conv1, conv2, conv3)
	//   ├── branchDialog1 (conv4, conv5) - 从conv2分叉
	//   └── branchDialog2 (conv6, conv7) - 从conv2分叉
	//     └── subBranchDialog (conv8, conv9) - 从conv6分叉

	t.Run("创建复杂分叉结构", func(t *testing.T) {
		// 创建第一个分叉 - 从conv2分叉
		conv2ID := int64(2)
		branch1DialogID, movedDialogID1, err := CreateBranchingDialogs(sessionID, conv2ID, rootDialogID)
		if err != nil {
			t.Fatalf("创建第一个分叉失败: %v", err)
		}

		// 在branch1Dialog中添加两个新conversations
		conv4 := models.ConversationModel{
			Prompt:    "问题4",
			Answer:    "回答4",
			SessionID: sessionID,
			DialogID:  branch1DialogID,
			Title:     "标题4",
			Summary:   "摘要4",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv4)
		
		conv5 := models.ConversationModel{
			Prompt:    "问题5",
			Answer:    "回答5",
			SessionID: sessionID,
			DialogID:  branch1DialogID,
			Title:     "标题5",
			Summary:   "摘要5",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv5)

		// 创建第二个分叉 - 也从conv2分叉
		branch2DialogID, _, err := CreateBranchingDialogs(sessionID, conv2ID, rootDialogID)
		if err != nil {
			t.Fatalf("创建第二个分叉失败: %v", err)
		}

		// 在branch2Dialog中添加两个新conversations
		conv6 := models.ConversationModel{
			Prompt:    "问题6",
			Answer:    "回答6",
			SessionID: sessionID,
			DialogID:  branch2DialogID,
			Title:     "标题6",
			Summary:   "摘要6",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv6)

		conv7 := models.ConversationModel{
			Prompt:    "问题7",
			Answer:    "回答7",
			SessionID: sessionID,
			DialogID:  branch2DialogID,
			Title:     "标题7",
			Summary:   "摘要7",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv7)

		// 创建子分叉 - 从conv6分叉
		subBranchDialogID, _, err := CreateBranchingDialogs(sessionID, conv6.ID, branch2DialogID)
		if err != nil {
			t.Fatalf("创建子分叉失败: %v", err)
		}

		// 在子分叉中添加两个conversations
		conv8 := models.ConversationModel{
			Prompt:    "问题8",
			Answer:    "回答8",
			SessionID: sessionID,
			DialogID:  subBranchDialogID,
			Title:     "标题8",
			Summary:   "摘要8",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv8)

		conv9 := models.ConversationModel{
			Prompt:    "问题9",
			Answer:    "回答9",
			SessionID: sessionID,
			DialogID:  subBranchDialogID,
			Title:     "标题9",
			Summary:   "摘要9",
			IsStarred: false,
			Comment:   "",
		}
		db.Create(&conv9)

		t.Logf("创建的复杂分叉结构:")
		t.Logf("  rootDialog: %d", rootDialogID)
		t.Logf("  branch1Dialog: %d", branch1DialogID)
		t.Logf("  movedDialog1: %d", movedDialogID1)
		t.Logf("  branch2Dialog: %d", branch2DialogID)
		t.Logf("  subBranchDialog: %d", subBranchDialogID)

		// 测试用例1：从子分叉中的conversation追溯父节点
		t.Run("子分叉conversation追溯父节点", func(t *testing.T) {
			conversations, err := traceParentConversationsFromConversation(conv9.ID, 5)
			if err != nil {
				t.Errorf("追溯父节点失败: %v", err)
			}

			t.Logf("从conv9追溯到的conversations数量: %d", len(conversations))
			for i, conv := range conversations {
				t.Logf("  [%d] Conv%d: %s (DialogID: %d)", i, conv.ID, conv.Prompt, conv.DialogID)
			}

			// 验证追溯路径中包含的prompt内容
			if len(conversations) < 3 {  // 至少应该追溯到几个关键节点
				t.Errorf("追溯路径长度太短，实际%d", len(conversations))
			}
			
			// 验证第一个是conv9
			if conversations[0].Prompt != "问题9" {
				t.Errorf("第一个节点应该是问题9，实际是%s", conversations[0].Prompt)
			}
		})

		// 测试用例2：从branch1中的conversation追溯父节点
		t.Run("branch1 conversation追溯父节点", func(t *testing.T) {
			conversations, err := traceParentConversationsFromConversation(conv5.ID, 4)
			if err != nil {
				t.Errorf("追溯父节点失败: %v", err)
			}

			t.Logf("从conv5追溯到的conversations数量: %d", len(conversations))
			for i, conv := range conversations {
				t.Logf("  [%d] Conv%d: %s (DialogID: %d)", i, conv.ID, conv.Prompt, conv.DialogID)
			}

			// 验证第一个是conv5
			if len(conversations) == 0 || conversations[0].Prompt != "问题5" {
				t.Errorf("第一个节点应该是问题5")
			}
		})

		// 测试用例3：从branch2中的conversation追溯父节点
		t.Run("branch2 conversation追溯父节点", func(t *testing.T) {
			conversations, err := traceParentConversationsFromConversation(conv7.ID, 4)
			if err != nil {
				t.Errorf("追溯父节点失败: %v", err)
			}

			t.Logf("从conv7追溯到的conversations数量: %d", len(conversations))
			for i, conv := range conversations {
				t.Logf("  [%d] Conv%d: %s (DialogID: %d)", i, conv.ID, conv.Prompt, conv.DialogID)
			}

			// 验证第一个是conv7
			if len(conversations) == 0 || conversations[0].Prompt != "问题7" {
				t.Errorf("第一个节点应该是问题7")
			}
		})

		// 测试用例4：测试findParentConversation函数在分叉点的行为
		t.Run("findParentConversation在分叉点的行为", func(t *testing.T) {
			// 找到conv8的父conversation（应该是conv6）
			parentConv, err := findParentConversation(conv8)
			if err != nil {
				t.Errorf("查找父conversation失败: %v", err)
			}

			t.Logf("Conv8的父conversation: Conv%d (Prompt: %s)", parentConv.ID, parentConv.Prompt)

			// 验证父conversation的内容包含"问题6"
			if parentConv.Prompt != "问题6" {
				t.Errorf("父conversation应该是问题6，实际是%s", parentConv.Prompt)
			}
		})

		// 测试用例5：测试上下文构建在复杂分叉场景下的正确性
		t.Run("复杂分叉场景下的上下文构建", func(t *testing.T) {
			context, err := BuildDialogContextFromConversation(sessionID, &conv9.ID, "新问题")
			if err != nil {
				t.Errorf("构建上下文失败: %v", err)
			}

			t.Logf("从conv9构建的上下文:\n%s", context)

			// 验证上下文中应该包含正确的追溯路径
			if !strings.Contains(context, "问题1") {
				t.Error("上下文应该包含问题1")
			}
			if !strings.Contains(context, "问题6") {
				t.Error("上下文应该包含问题6")
			}
			if !strings.Contains(context, "问题9") {
				t.Error("上下文应该包含问题9")
			}

			// 不应该包含其他分支的内容
			if strings.Contains(context, "问题4") || strings.Contains(context, "问题5") {
				t.Error("上下文不应该包含其他分支的内容（问题4或问题5）")
			}
		})

		// 测试用例6：验证不同分支的上下文隔离
		t.Run("验证分支间上下文隔离", func(t *testing.T) {
			// 测试branch1的上下文
			context1, err := BuildDialogContextFromConversation(sessionID, &conv5.ID, "新问题")
			if err != nil {
				t.Errorf("构建branch1上下文失败: %v", err)
			}

			// 测试branch2的上下文
			context2, err := BuildDialogContextFromConversation(sessionID, &conv7.ID, "新问题")
			if err != nil {
				t.Errorf("构建branch2上下文失败: %v", err)
			}

			t.Logf("Branch1上下文:\n%s", context1)
			t.Logf("Branch2上下文:\n%s", context2)

			// branch1的上下文不应该包含branch2的内容
			if strings.Contains(context1, "问题6") || strings.Contains(context1, "问题7") {
				t.Error("Branch1上下文不应该包含Branch2的内容")
			}

			// branch2的上下文不应该包含branch1的内容
			if strings.Contains(context2, "问题4") || strings.Contains(context2, "问题5") {
				t.Error("Branch2上下文不应该包含Branch1的内容")
			}

			// 两个上下文都应该包含共同的祖先
			if !strings.Contains(context1, "问题1") {
				t.Error("Branch1上下文应该包含共同祖先问题1")
			}
			if !strings.Contains(context2, "问题1") {
				t.Error("Branch2上下文应该包含共同祖先问题1")
			}
		})
	})
}