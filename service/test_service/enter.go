// Path: ./service/test_service/enter.go

package test_service

import (
	"dialogTree/conf"
	"dialogTree/global"
	"dialogTree/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

// SetupTestEnvironment 设置测试环境
func SetupTestEnvironment(t *testing.T) (*gorm.DB, *gin.Engine) {
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

	return db, router
}
