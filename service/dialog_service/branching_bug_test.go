package dialog_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestBranchingBug 测试分叉bug：从中间节点分叉时，原有的父子关系被错误地修改
// 场景：c1->c2, c2->c3, c2->c4，从c1分叉后应该保持原关系，但实际变成了c1->c2, c1->c3, c1->c4
func TestBranchingBug(t *testing.T) {
	// 设置测试环境
	setupTestConfig()
	db := setupTestDB(t)
	global.DB = db

	// 创建基础数据
	sessionID, dialogID := createBasicTestData(t, db)

	// 创建对话链：c1->c2->c3, c1->c2->c4
	// 首先创建c1, c2
	c1 := createConversation(t, db, sessionID, dialogID, "问题1", "回答1", time.Now())
	c2 := createConversation(t, db, sessionID, dialogID, "问题2", "回答2", time.Now().Add(1*time.Minute))

	// 从c2分叉创建c3和c4的分支
	needsBranching, err := CheckIfBranchingByConversation(c2.ID)
	if err != nil {
		t.Fatalf("检查分叉失败: %v", err)
	}
	if needsBranching {
		t.Fatal("c2应该是最新的，不需要分叉")
	}

	// 在c2后面创建c3
	c3 := createConversation(t, db, sessionID, dialogID, "问题3", "回答3", time.Now().Add(2*time.Minute))

	// 现在从c2分叉创建c4（这应该会触发分叉逻辑）
	needsBranching, err = CheckIfBranchingByConversation(c2.ID)
	if err != nil {
		t.Fatalf("检查分叉失败: %v", err)
	}
	if !needsBranching {
		t.Fatal("从c2分叉应该需要分叉，因为c3已经存在")
	}

	// 执行分叉（这里会暴露bug）
	newDialogID, branchedDialogID, err := CreateBranchingDialogs(sessionID, c2.ID, dialogID)
	if err != nil {
		t.Fatalf("创建分叉失败: %v", err)
	}

	// 在新的dialog中创建c4
	c4 := createConversation(t, db, sessionID, newDialogID, "问题4", "回答4", time.Now().Add(3*time.Minute))

	t.Run("验证分叉后的数据结构", func(t *testing.T) {
		// 查看原始dialog中剩余的conversations
		var originalConversations []models.ConversationModel
		db.Where("dialog_id = ?", dialogID).Order("created_at ASC").Find(&originalConversations)

		t.Logf("原始dialog中的conversations数量: %d", len(originalConversations))
		for i, conv := range originalConversations {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 查看被分叉出去的dialog中的conversations
		var branchedConversations []models.ConversationModel
		db.Where("dialog_id = ?", branchedDialogID).Order("created_at ASC").Find(&branchedConversations)

		t.Logf("分叉dialog中的conversations数量: %d", len(branchedConversations))
		for i, conv := range branchedConversations {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 查看新dialog中的conversations
		var newConversations []models.ConversationModel
		db.Where("dialog_id = ?", newDialogID).Order("created_at ASC").Find(&newConversations)

		t.Logf("新dialog中的conversations数量: %d", len(newConversations))
		for i, conv := range newConversations {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 这里会暴露bug：c3被错误地移动到了分叉dialog中
		// 期望的结果：
		// - 原始dialog: c1, c2
		// - 分叉dialog: c3 (保持c2->c3的关系)
		// - 新dialog: c4

		// 验证原始dialog应该只有c1和c2
		if len(originalConversations) != 2 {
			t.Errorf("原始dialog应该有2个conversations（c1,c2），实际有%d个", len(originalConversations))
		}

		// 验证分叉dialog应该只有c3
		if len(branchedConversations) != 1 {
			t.Errorf("分叉dialog应该有1个conversation（c3），实际有%d个", len(branchedConversations))
		}

		// 验证新dialog应该有c4
		if len(newConversations) != 1 {
			t.Errorf("新dialog应该有1个conversation（c4），实际有%d个", len(newConversations))
		}
	})

	t.Run("验证上下文追溯的正确性", func(t *testing.T) {
		// 从c3追溯父节点应该能找到c2, c1
		c3Ancestors, err := traceParentConversationsFromConversation(c3.ID, 5)
		if err != nil {
			t.Errorf("从c3追溯父节点失败: %v", err)
		}

		t.Logf("从c3追溯到的ancestors数量: %d", len(c3Ancestors))
		for i, conv := range c3Ancestors {
			t.Logf("  [%d] %s (ID: %d, DialogID: %d)", i, conv.Prompt, conv.ID, conv.DialogID)
		}

		// 验证追溯路径：c3 -> c2 -> c1
		if len(c3Ancestors) < 3 {
			t.Errorf("c3的追溯路径应该包含至少3个节点（c3,c2,c1），实际%d个", len(c3Ancestors))
		}

		// 验证第一个是c3本身
		if c3Ancestors[0].ID != c3.ID {
			t.Errorf("追溯路径第一个应该是c3，实际是%d", c3Ancestors[0].ID)
		}

		// 验证能追溯到c2
		found_c2 := false
		for _, conv := range c3Ancestors {
			if conv.ID == c2.ID {
				found_c2 = true
				break
			}
		}
		if !found_c2 {
			t.Error("从c3追溯应该能找到c2")
		}

		// 验证能追溯到c1
		found_c1 := false
		for _, conv := range c3Ancestors {
			if conv.ID == c1.ID {
				found_c1 = true
				break
			}
		}
		if !found_c1 {
			t.Error("从c3追溯应该能找到c1")
		}

		// 从c4追溯父节点也应该能找到c2, c1
		c4Ancestors, err := traceParentConversationsFromConversation(c4.ID, 5)
		if err != nil {
			t.Errorf("从c4追溯父节点失败: %v", err)
		}

		t.Logf("从c4追溯到的ancestors数量: %d", len(c4Ancestors))
		for i, conv := range c4Ancestors {
			t.Logf("  [%d] %s (ID: %d, DialogID: %d)", i, conv.Prompt, conv.ID, conv.DialogID)
		}

		// 验证c4也能追溯到c2和c1
		found_c2_from_c4 := false
		found_c1_from_c4 := false
		for _, conv := range c4Ancestors {
			if conv.ID == c2.ID {
				found_c2_from_c4 = true
			}
			if conv.ID == c1.ID {
				found_c1_from_c4 = true
			}
		}
		if !found_c2_from_c4 {
			t.Error("从c4追溯应该能找到c2")
		}
		if !found_c1_from_c4 {
			t.Error("从c4追溯应该能找到c1")
		}
	})

	// 演示预期的正确行为
	t.Run("演示预期的正确行为", func(t *testing.T) {
		t.Log("=== 预期的正确分叉行为 ===")
		t.Log("原始结构: c1->c2->c3")
		t.Log("从c1分叉后应该是:")
		t.Log("  原始分支: c1->c2->c3")
		t.Log("  新分支: c1->c4")
		t.Log("")
		t.Log("但实际上由于bug，变成了:")
		t.Log("  原始分支: c1->c2")
		t.Log("  分叉分支: c1->c3 (错误！应该是c2->c3)")
		t.Log("  新分支: c1->c4")
		t.Log("")
		t.Log("这破坏了c2->c3的父子关系，c3错误地变成了c1的直接子节点")
	})
}

// createBasicTestData 创建基础测试数据
func createBasicTestData(t *testing.T, db *gorm.DB) (int64, int64) {
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

	return session.ID, dialog.ID
}

// createConversation 创建一个conversation
func createConversation(t *testing.T, db *gorm.DB, sessionID, dialogID int64, prompt, answer string, createdAt time.Time) *models.ConversationModel {
	conv := &models.ConversationModel{
		Prompt:    prompt,
		Answer:    answer,
		SessionID: sessionID,
		DialogID:  dialogID,
		Title:     fmt.Sprintf("标题-%s", prompt),
		Summary:   fmt.Sprintf("摘要-%s", answer),
		IsStarred: false,
		Comment:   "",
	}
	conv.CreatedAt = createdAt
	conv.UpdatedAt = createdAt

	err := db.Create(conv).Error
	if err != nil {
		t.Fatalf("创建conversation失败: %v", err)
	}

	return conv
}

