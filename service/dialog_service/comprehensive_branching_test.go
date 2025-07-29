package dialog_service

import (
	"dialogTree/conf"
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestComprehensiveBranchingLogic 全面测试分叉逻辑
// 这个测试覆盖了CreateBranchingDialogs函数的所有关键功能，包括：
// 1. 创建新dialog用于新分支
// 2. 创建分叉dialog用于分叉出的对话
// 3. 将原父dialog的子dialogs重新指向分叉dialog (第595-602行的核心逻辑)
// 4. 移动分叉点之后的conversations
// 5. 正确建立分支关系和分叉点记录
func TestComprehensiveBranchingLogic(t *testing.T) {
	// 设置测试环境
	setupComprehensiveTestConfig()
	db := setupComprehensiveTestDB(t)
	global.DB = db

	// 创建复杂的测试场景
	sessionID, rootDialogID := setupComplexBranchingScenario(t, db)

	t.Logf("=== 初始复杂场景设置完成 ===")
	t.Logf("SessionID: %d, RootDialogID: %d", sessionID, rootDialogID)

	// 验证初始状态
	validateInitialState(t, db, sessionID, rootDialogID)

	// 执行分叉测试
	t.Run("从中间节点分叉", func(t *testing.T) {
		// 选择root dialog中的第2个conversation作为分叉点
		var branchPointConv models.ConversationModel
		err := db.Where("dialog_id = ?", rootDialogID).
			Order("created_at ASC").
			Offset(1).
			First(&branchPointConv).Error
		if err != nil {
			t.Fatalf("获取分叉点conversation失败: %v", err)
		}

		t.Logf("选择分叉点: %s (ID: %d)", branchPointConv.Prompt, branchPointConv.ID)

		// 检查是否需要分叉
		needsBranching, err := CheckIfBranchingByConversation(branchPointConv.ID)
		if err != nil {
			t.Fatalf("检查分叉失败: %v", err)
		}

		if !needsBranching {
			t.Fatal("应该需要分叉")
		}

		// 执行分叉
		newDialogID, branchedDialogID, err := CreateBranchingDialogs(sessionID, branchPointConv.ID, rootDialogID)
		if err != nil {
			t.Fatalf("创建分叉失败: %v", err)
		}

		t.Logf("分叉完成: newDialogID=%d, branchedDialogID=%d", newDialogID, branchedDialogID)

		// 验证分叉结果
		validateBranchingResults(t, db, sessionID, rootDialogID, newDialogID, branchedDialogID, branchPointConv.ID)
	})

	// 测试更复杂的多层分叉
	t.Run("多层分叉测试", func(t *testing.T) {
		// 在已有的分叉基础上再进行分叉
		testMultiLevelBranching(t, db, sessionID)
	})

	// 测试分叉后的上下文追溯
	t.Run("分叉后上下文追溯", func(t *testing.T) {
		testContextTracingAfterBranching(t, db, sessionID)
	})
}

// setupComprehensiveTestConfig 设置全面测试配置
func setupComprehensiveTestConfig() {
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

// setupComprehensiveTestDB 设置全面测试数据库
func setupComprehensiveTestDB(t *testing.T) *gorm.DB {
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

// setupComplexBranchingScenario 创建复杂的分叉测试场景
func setupComplexBranchingScenario(t *testing.T, db *gorm.DB) (int64, int64) {
	// 创建类别
	category := models.CategoryModel{Name: "分叉测试类别"}
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	db.Create(&category)

	// 创建会话
	session := models.SessionModel{
		Tittle:     "分叉测试会话",
		Summary:    "分叉测试摘要",
		CategoryID: category.ID,
	}
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	db.Create(&session)

	// 创建根对话树
	rootDialog := models.DialogModel{
		SessionID: session.ID,
		ParentID:  nil,
	}
	rootDialog.CreatedAt = time.Now()
	rootDialog.UpdatedAt = time.Now()
	db.Create(&rootDialog)

	// 更新session
	session.RootDialogID = &rootDialog.ID
	db.Save(&session)

	// 在根dialog中创建5个conversations: c1->c2->c3->c4->c5
	for i := 1; i <= 5; i++ {
		conv := models.ConversationModel{
			Prompt:    fmt.Sprintf("根对话问题%d", i),
			Answer:    fmt.Sprintf("根对话回答%d", i),
			SessionID: session.ID,
			DialogID:  rootDialog.ID,
			Title:     fmt.Sprintf("根对话标题%d", i),
			Summary:   fmt.Sprintf("根对话摘要%d", i),
		}
		conv.CreatedAt = time.Now().Add(time.Duration(i-1) * time.Minute)
		conv.UpdatedAt = conv.CreatedAt
		db.Create(&conv)
	}

	// 创建已有的子dialogs（模拟之前的分叉）
	existingChildDialog1 := models.DialogModel{
		SessionID: session.ID,
		ParentID:  &rootDialog.ID,
	}
	existingChildDialog1.CreatedAt = time.Now().Add(10 * time.Minute)
	existingChildDialog1.UpdatedAt = existingChildDialog1.CreatedAt
	db.Create(&existingChildDialog1)

	existingChildDialog2 := models.DialogModel{
		SessionID: session.ID,
		ParentID:  &rootDialog.ID,
	}
	existingChildDialog2.CreatedAt = time.Now().Add(15 * time.Minute)
	existingChildDialog2.UpdatedAt = existingChildDialog2.CreatedAt
	db.Create(&existingChildDialog2)

	// 在现有子dialogs中添加conversations
	for i := 1; i <= 2; i++ {
		conv1 := models.ConversationModel{
			Prompt:    fmt.Sprintf("子对话1-问题%d", i),
			Answer:    fmt.Sprintf("子对话1-回答%d", i),
			SessionID: session.ID,
			DialogID:  existingChildDialog1.ID,
			Title:     fmt.Sprintf("子对话1-标题%d", i),
			Summary:   fmt.Sprintf("子对话1-摘要%d", i),
		}
		conv1.CreatedAt = time.Now().Add(time.Duration(10+i) * time.Minute)
		conv1.UpdatedAt = conv1.CreatedAt
		db.Create(&conv1)

		conv2 := models.ConversationModel{
			Prompt:    fmt.Sprintf("子对话2-问题%d", i),
			Answer:    fmt.Sprintf("子对话2-回答%d", i),
			SessionID: session.ID,
			DialogID:  existingChildDialog2.ID,
			Title:     fmt.Sprintf("子对话2-标题%d", i),
			Summary:   fmt.Sprintf("子对话2-摘要%d", i),
		}
		conv2.CreatedAt = time.Now().Add(time.Duration(15+i) * time.Minute)
		conv2.UpdatedAt = conv2.CreatedAt
		db.Create(&conv2)
	}

	return session.ID, rootDialog.ID
}

// validateInitialState 验证初始状态
func validateInitialState(t *testing.T, db *gorm.DB, sessionID, rootDialogID int64) {
	// 验证根dialog中有5个conversations
	var rootConvs []models.ConversationModel
	db.Where("dialog_id = ?", rootDialogID).Order("created_at ASC").Find(&rootConvs)
	if len(rootConvs) != 5 {
		t.Fatalf("根dialog应该有5个conversations，实际有%d个", len(rootConvs))
	}

	// 验证有2个子dialogs
	var childDialogs []models.DialogModel
	db.Where("parent_id = ?", rootDialogID).Find(&childDialogs)
	if len(childDialogs) != 2 {
		t.Fatalf("应该有2个子dialogs，实际有%d个", len(childDialogs))
	}

	t.Logf("初始状态验证通过: 根dialog有%d个conversations, %d个子dialogs", len(rootConvs), len(childDialogs))
}

// validateBranchingResults 验证分叉结果
func validateBranchingResults(t *testing.T, db *gorm.DB, sessionID, originalRootDialogID, newDialogID, branchedDialogID, branchPointConvID int64) {
	t.Logf("=== 开始验证分叉结果 ===")

	// 1. 验证新dialogs的创建和属性
	var newDialog, branchedDialog models.DialogModel
	
	err := db.First(&newDialog, newDialogID).Error
	if err != nil {
		t.Fatalf("新dialog不存在: %v", err)
	}
	
	err = db.First(&branchedDialog, branchedDialogID).Error
	if err != nil {
		t.Fatalf("分叉dialog不存在: %v", err)
	}

	// 验证父子关系
	if newDialog.ParentID == nil || *newDialog.ParentID != originalRootDialogID {
		t.Errorf("新dialog的父ID应该是%d，实际是%v", originalRootDialogID, newDialog.ParentID)
	}
	
	if branchedDialog.ParentID == nil || *branchedDialog.ParentID != originalRootDialogID {
		t.Errorf("分叉dialog的父ID应该是%d，实际是%v", originalRootDialogID, branchedDialog.ParentID)
	}

	// 验证分叉点记录
	if newDialog.BranchFromConversationID == nil || *newDialog.BranchFromConversationID != branchPointConvID {
		t.Errorf("新dialog的分叉点应该是%d，实际是%v", branchPointConvID, newDialog.BranchFromConversationID)
	}
	
	if branchedDialog.BranchFromConversationID == nil || *branchedDialog.BranchFromConversationID != branchPointConvID {
		t.Errorf("分叉dialog的分叉点应该是%d，实际是%v", branchPointConvID, branchedDialog.BranchFromConversationID)
	}

	t.Logf("✅ Dialog创建和属性验证通过")

	// 2. 验证conversations的移动
	// 原dialog应该只有分叉点及之前的conversations
	var originalConvs []models.ConversationModel
	db.Where("dialog_id = ?", originalRootDialogID).Order("created_at ASC").Find(&originalConvs)
	
	// 分叉dialog应该有分叉点之后的conversations
	var branchedConvs []models.ConversationModel
	db.Where("dialog_id = ?", branchedDialogID).Order("created_at ASC").Find(&branchedConvs)
	
	// 新dialog应该为空（用于新对话）
	var newConvs []models.ConversationModel
	db.Where("dialog_id = ?", newDialogID).Find(&newConvs)

	t.Logf("原dialog剩余conversations: %d个", len(originalConvs))
	t.Logf("分叉dialog包含conversations: %d个", len(branchedConvs))
	t.Logf("新dialog包含conversations: %d个", len(newConvs))

	// 验证分叉点位置
	var branchPointConv models.ConversationModel
	db.First(&branchPointConv, branchPointConvID)
	
	// 原dialog应该包含分叉点及之前的conversations
	foundBranchPoint := false
	for _, conv := range originalConvs {
		if conv.ID == branchPointConvID {
			foundBranchPoint = true
			break
		}
	}
	if !foundBranchPoint {
		t.Error("原dialog应该包含分叉点conversation")
	}

	// 分叉dialog应该包含分叉点之后的conversations
	if len(branchedConvs) == 0 {
		t.Error("分叉dialog应该包含分叉点之后的conversations")
	}

	// 新dialog应该为空
	if len(newConvs) != 0 {
		t.Errorf("新dialog应该为空，实际有%d个conversations", len(newConvs))
	}

	t.Logf("✅ Conversations移动验证通过")

	// 3. 验证子dialogs的重新指向（这是595-602行的核心逻辑）
	var updatedChildDialogs []models.DialogModel
	db.Where("parent_id = ?", branchedDialogID).Find(&updatedChildDialogs)
	
	var remainingChildDialogs []models.DialogModel
	db.Where("parent_id = ? AND id NOT IN ?", originalRootDialogID, []int64{newDialogID, branchedDialogID}).Find(&remainingChildDialogs)

	t.Logf("被重新指向到分叉dialog的子dialogs: %d个", len(updatedChildDialogs))
	t.Logf("仍指向原dialog的子dialogs: %d个", len(remainingChildDialogs))

	// 原来的子dialogs应该被重新指向分叉dialog
	if len(updatedChildDialogs) != 2 {
		t.Errorf("应该有2个子dialogs被重新指向分叉dialog，实际有%d个", len(updatedChildDialogs))
	}

	// 除了新创建的两个dialog，原dialog不应该有其他子dialogs
	if len(remainingChildDialogs) != 0 {
		t.Errorf("原dialog不应该有其他子dialogs，实际有%d个", len(remainingChildDialogs))
	}

	t.Logf("✅ 子dialogs重新指向验证通过")

	// 4. 验证完整的dialog树结构
	validateDialogTreeStructure(t, db, sessionID, originalRootDialogID, newDialogID, branchedDialogID)

	t.Logf("=== 分叉结果验证完成 ===")
}

// validateDialogTreeStructure 验证dialog树结构
func validateDialogTreeStructure(t *testing.T, db *gorm.DB, sessionID, rootDialogID, newDialogID, branchedDialogID int64) {
	t.Logf("=== 验证Dialog树结构 ===")

	// 获取所有dialogs
	var allDialogs []models.DialogModel
	db.Where("session_id = ?", sessionID).Find(&allDialogs)

	t.Logf("总共有%d个dialogs:", len(allDialogs))
	for _, dialog := range allDialogs {
		parentInfo := "nil"
		if dialog.ParentID != nil {
			parentInfo = fmt.Sprintf("%d", *dialog.ParentID)
		}
		branchInfo := "nil"
		if dialog.BranchFromConversationID != nil {
			branchInfo = fmt.Sprintf("%d", *dialog.BranchFromConversationID)
		}
		t.Logf("  Dialog %d: ParentID=%s, BranchFrom=%s", dialog.ID, parentInfo, branchInfo)
	}

	// 验证树结构的正确性
	// 1. 根dialog
	rootCount := 0
	// 2. 根dialog的直接子dialogs
	rootChildCount := 0
	// 3. 分叉dialog的子dialogs
	branchedChildCount := 0

	for _, dialog := range allDialogs {
		if dialog.ParentID == nil {
			rootCount++
		} else if *dialog.ParentID == rootDialogID {
			rootChildCount++
		} else if *dialog.ParentID == branchedDialogID {
			branchedChildCount++
		}
	}

	if rootCount != 1 {
		t.Errorf("应该有1个根dialog，实际有%d个", rootCount)
	}

	if rootChildCount != 2 { // newDialog + branchedDialog
		t.Errorf("根dialog应该有2个直接子dialogs，实际有%d个", rootChildCount)
	}

	if branchedChildCount != 2 { // 原来的2个子dialogs被重新指向
		t.Errorf("分叉dialog应该有2个子dialogs，实际有%d个", branchedChildCount)
	}

	t.Logf("✅ Dialog树结构验证通过")
}

// testMultiLevelBranching 测试多层分叉
func testMultiLevelBranching(t *testing.T, db *gorm.DB, sessionID int64) {
	t.Logf("=== 测试多层分叉 ===")

	// 在已有的分叉dialog中再进行分叉
	// 这里可以添加更复杂的多层分叉测试逻辑
	var branchedDialogs []models.DialogModel
	db.Where("session_id = ? AND parent_id IS NOT NULL AND branch_from_conversation_id IS NOT NULL", sessionID).Find(&branchedDialogs)

	if len(branchedDialogs) > 0 {
		// 选择第一个分叉dialog进行进一步测试
		branchedDialog := branchedDialogs[0]
		
		// 在分叉dialog中添加一些conversations然后再分叉
		for i := 1; i <= 3; i++ {
			conv := models.ConversationModel{
				Prompt:    fmt.Sprintf("二级分叉问题%d", i),
				Answer:    fmt.Sprintf("二级分叉回答%d", i),
				SessionID: sessionID,
				DialogID:  branchedDialog.ID,
				Title:     fmt.Sprintf("二级分叉标题%d", i),
				Summary:   fmt.Sprintf("二级分叉摘要%d", i),
			}
			conv.CreatedAt = time.Now().Add(time.Duration(30+i) * time.Minute)
			conv.UpdatedAt = conv.CreatedAt
			db.Create(&conv)
		}

		// 从第2个conversation进行二级分叉
		var secondLevelBranchPoint models.ConversationModel
		err := db.Where("dialog_id = ?", branchedDialog.ID).
			Order("created_at ASC").
			Offset(1).
			First(&secondLevelBranchPoint).Error
		
		if err == nil {
			needsBranching, err := CheckIfBranchingByConversation(secondLevelBranchPoint.ID)
			if err == nil && needsBranching {
				newDialogID, newBranchedDialogID, err := CreateBranchingDialogs(sessionID, secondLevelBranchPoint.ID, branchedDialog.ID)
				if err != nil {
					t.Errorf("二级分叉失败: %v", err)
				} else {
					t.Logf("二级分叉成功: newDialogID=%d, branchedDialogID=%d", newDialogID, newBranchedDialogID)
				}
			}
		}
	}

	t.Logf("✅ 多层分叉测试完成")
}

// testContextTracingAfterBranching 测试分叉后的上下文追溯
func testContextTracingAfterBranching(t *testing.T, db *gorm.DB, sessionID int64) {
	t.Logf("=== 测试分叉后的上下文追溯 ===")

	// 获取所有conversations
	var allConversations []models.ConversationModel
	db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&allConversations)

	// 测试从不同分支的conversations进行上下文追溯
	for i, conv := range allConversations {
		if i >= 3 { // 只测试几个代表性的conversations
			break
		}

		ancestors, err := traceParentConversationsFromConversation(conv.ID, 10)
		if err != nil {
			t.Errorf("从conversation %d追溯失败: %v", conv.ID, err)
			continue
		}

		t.Logf("从conversation %d (%s) 追溯到 %d 个ancestors:", conv.ID, conv.Prompt, len(ancestors))
		for j, ancestor := range ancestors {
			t.Logf("  [%d] %s (ID: %d, DialogID: %d)", j, ancestor.Prompt, ancestor.ID, ancestor.DialogID)
		}
	}

	t.Logf("✅ 上下文追溯测试完成")
}