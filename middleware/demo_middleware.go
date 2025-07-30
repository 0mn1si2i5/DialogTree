// Path: ./middleware/demo_middleware.go

package middleware

import (
	"dialogTree/global"
	"dialogTree/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	demoTicker  *time.Ticker
	demoRunning bool
	demoMutex   sync.Mutex
)

func DemoMiddleware(c *gin.Context) {
	if !global.Config.System.Demo {
		return
	}
	demoMutex.Lock()
	defer demoMutex.Unlock()

	// 如果timer没有在运行，则启动新的timer
	if !demoRunning {
		startDemoTimer()
	}
}

func startDemoTimer() {
	hours := time.Duration(global.Config.System.DemoTimer)
	demoTicker = time.NewTicker(hours * time.Hour)
	demoRunning = true

	logrus.Infof("Demo timer started, DB will be reset in %d hours.", hours)

	go func() {
		for range demoTicker.C {
			logrus.Info("Demo timer triggered, resetting database...")
			initDB()
			return // 重置后退出goroutine
		}
	}()
}

// StopDemoTimer 停止demo定时器 (可选，用于优雅关闭)
func StopDemoTimer() {
	if demoTicker != nil {
		demoTicker.Stop()
		logrus.Info("Demo timer stopped.")
	}
}

// 初始化数据库
func initDB() {
	demoMutex.Lock()
	defer demoMutex.Unlock()

	// 1. 清空所有表（注意外键顺序）
	tables := []string{
		"conversation_models", // 先删除子表
		"dialog_models",
		"session_models",
		"category_models", // 最后删除父表
	}

	for _, table := range tables {
		if err := global.DB.Exec("TRUNCATE TABLE " +
			table).Error; err != nil {
			logrus.Errorf("Failed to truncate table %s: %v",
				table, err)
			return
		}
	}

	// 2. 插入样板数据
	insertSampleData()

	// 3. 停止timer
	if demoTicker != nil {
		demoTicker.Stop()
		demoTicker = nil
	}
	demoRunning = false

	logrus.Info("Database reset completed with sample data. Demo timer stopped.")
}

func insertSampleData() {
	// 创建样板分类
	category := models.CategoryModel{Name: "Demo"}
	global.DB.Create(&category)

	// 创建样板会话
	session := models.SessionModel{
		Tittle:     "Welcome to DialogTree Demo",
		Summary:    "This is a sample conversation",
		CategoryID: category.ID,
	}
	global.DB.Create(&session)

	// 创建样板对话等...
}
