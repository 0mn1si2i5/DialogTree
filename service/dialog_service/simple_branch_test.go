package dialog_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"testing"
	"time"
)

// TestSimpleBranchingScenario 测试简单的分叉场景
func TestSimpleBranchingScenario(t *testing.T) {
	// 设置测试环境
	setupTestConfig()
	db := setupTestDB(t)
	global.DB = db

	// 创建基础数据
	sessionID, dialogID := createBasicTestData(t, db)

	// 创建对话链：d1有c1->c2->c3
	c1 := createConversation(t, db, sessionID, dialogID, "问题1", "回答1", time.Now())
	c2 := createConversation(t, db, sessionID, dialogID, "问题2", "回答2", time.Now().Add(1*time.Minute))
	c3 := createConversation(t, db, sessionID, dialogID, "问题3", "回答3", time.Now().Add(2*time.Minute))

	t.Logf("初始状态:")
	t.Logf("  d1: c1->c2->c3")
	t.Logf("  c1 ID: %d, c2 ID: %d, c3 ID: %d", c1.ID, c2.ID, c3.ID)

	// 选择c2作为父节点进行分叉
	t.Log("\n选择c2作为父节点进行分叉...")

	// 检查是否需要分叉
	needsBranching, err := CheckIfBranchingByConversation(c2.ID)
	if err != nil {
		t.Fatalf("检查分叉失败: %v", err)
	}
	if !needsBranching {
		t.Fatal("从c2分叉应该需要分叉，因为c3已经存在")
	}

	// 执行分叉
	newDialogID, branchedDialogID, err := CreateBranchingDialogs(sessionID, c2.ID, dialogID)
	if err != nil {
		t.Fatalf("创建分叉失败: %v", err)
	}

	t.Logf("分叉后:")
	t.Logf("  原始dialogID: %d", dialogID)
	t.Logf("  新dialogID: %d", newDialogID)
	t.Logf("  分叉dialogID: %d", branchedDialogID)

	// 检查各个dialog中的conversations
	t.Run("验证分叉后的数据分布", func(t *testing.T) {
		// 查看原始dialog中的conversations
		var originalConvs []models.ConversationModel
		db.Where("dialog_id = ?", dialogID).Order("created_at ASC").Find(&originalConvs)

		t.Logf("原始dialog(%d)中的conversations:", dialogID)
		for i, conv := range originalConvs {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 查看分叉dialog中的conversations
		var branchedConvs []models.ConversationModel
		db.Where("dialog_id = ?", branchedDialogID).Order("created_at ASC").Find(&branchedConvs)

		t.Logf("分叉dialog(%d)中的conversations:", branchedDialogID)
		for i, conv := range branchedConvs {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 查看新dialog中的conversations（应该为空）
		var newConvs []models.ConversationModel
		db.Where("dialog_id = ?", newDialogID).Order("created_at ASC").Find(&newConvs)

		t.Logf("新dialog(%d)中的conversations:", newDialogID)
		for i, conv := range newConvs {
			t.Logf("  [%d] %s (ID: %d)", i, conv.Prompt, conv.ID)
		}

		// 验证期望的结果
		// 根据用户描述：
		// - d1: c1->c2 (保留到分叉点)
		// - d2: c3 (c3被移动到新dialog)
		// - d3: 空 (用于新对话)

		if len(originalConvs) != 2 {
			t.Errorf("原始dialog应该有2个conversations(c1,c2)，实际有%d个", len(originalConvs))
		} else {
			if originalConvs[0].ID != c1.ID || originalConvs[1].ID != c2.ID {
				t.Error("原始dialog应该包含c1和c2")
			}
		}

		if len(branchedConvs) != 1 {
			t.Errorf("分叉dialog应该有1个conversation(c3)，实际有%d个", len(branchedConvs))
		} else {
			if branchedConvs[0].ID != c3.ID {
				t.Error("分叉dialog应该包含c3")
			}
		}

		if len(newConvs) != 0 {
			t.Errorf("新dialog应该为空，实际有%d个conversations", len(newConvs))
		}
	})

	// 现在在新dialog中添加新对话
	t.Run("在新dialog中添加新对话", func(t *testing.T) {
		c4 := createConversation(t, db, sessionID, newDialogID, "问题4", "回答4", time.Now().Add(3*time.Minute))

		t.Logf("添加新对话后:")
		t.Logf("  d1: c1->c2")
		t.Logf("  d2: c3")
		t.Logf("  d3: c4")

		// 验证上下文追溯
		// 从c3追溯应该能找到c2, c1
		c3Ancestors, err := traceParentConversationsFromConversation(c3.ID, 5)
		if err != nil {
			t.Errorf("从c3追溯失败: %v", err)
		}

		t.Logf("从c3追溯的路径:")
		for i, conv := range c3Ancestors {
			t.Logf("  [%d] %s (ID: %d, DialogID: %d)", i, conv.Prompt, conv.ID, conv.DialogID)
		}

		// 从c4追溯应该能找到c2, c1
		c4Ancestors, err := traceParentConversationsFromConversation(c4.ID, 5)
		if err != nil {
			t.Errorf("从c4追溯失败: %v", err)
		}

		t.Logf("从c4追溯的路径:")
		for i, conv := range c4Ancestors {
			t.Logf("  [%d] %s (ID: %d, DialogID: %d)", i, conv.Prompt, conv.ID, conv.DialogID)
		}

		// 验证追溯结果
		// c3应该能追溯到: c3 -> c2 -> c1
		if len(c3Ancestors) < 3 {
			t.Errorf("c3应该能追溯到3个节点(c3,c2,c1)，实际%d个", len(c3Ancestors))
		}

		// c4应该能追溯到: c4 -> c2 -> c1
		if len(c4Ancestors) < 3 {
			t.Errorf("c4应该能追溯到3个节点(c4,c2,c1)，实际%d个", len(c4Ancestors))
		}

		// 验证c3和c4都能找到c2作为分叉点
		foundC2FromC3 := false
		foundC2FromC4 := false
		
		for _, conv := range c3Ancestors {
			if conv.ID == c2.ID {
				foundC2FromC3 = true
				break
			}
		}
		
		for _, conv := range c4Ancestors {
			if conv.ID == c2.ID {
				foundC2FromC4 = true
				break
			}
		}

		if !foundC2FromC3 {
			t.Error("c3应该能追溯到分叉点c2")
		}
		if !foundC2FromC4 {
			t.Error("c4应该能追溯到分叉点c2")
		}
	})
}