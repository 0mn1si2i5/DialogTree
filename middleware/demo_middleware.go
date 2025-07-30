// Path: ./middleware/demo_middleware.go

package middleware

import (
	"dialogTree/global"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
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

func TestDBRestarter() {
	initDB()
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
	logrus.Info("Database truncated successfully")

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
	sqlFile := "sample_data.sql"
	logrus.Infof("Inserting sample data from %s", sqlFile)

	// 分割多条SQL语句
	sqlStatements, err := washMySQLDump(sqlFile)
	if err != nil || sqlStatements == nil {
		logrus.Errorf("Failed to wash SQL file: %v", err)
		return
	}
	if len(sqlStatements) == 0 {
		logrus.Error("No SQL statements found in SQL file")
		return
	}
	logrus.Debugf("Found %d SQL statements in SQL file", len(sqlStatements))

	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := global.DB.Exec(stmt).Error; err != nil {
			logrus.Errorf("Failed to execute SQL statement: %v", err)
			logrus.Errorf("Statement: %s", stmt)
			return
		}
	}
	logrus.Info("Sample data inserted successfully")
}

func washMySQLDump(path string) ([]string, error) {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rawSQL := string(sqlBytes)

	// 清理 /*! MySQL特殊语句 */
	reSpecial := regexp.MustCompile(`(?s)/\*![0-9]{5}.*?\*/`)
	rawSQL = reSpecial.ReplaceAllString(rawSQL, "")

	// 清理 LOCK / UNLOCK
	reLock := regexp.MustCompile(`(?m)^\s*(LOCK TABLES|UNLOCK TABLES).*?\n`)
	rawSQL = reLock.ReplaceAllString(rawSQL, "")

	// 注释掉 ALTER TABLE ... KEYS
	reAlter := regexp.MustCompile(`(?im)^\s*ALTER TABLE\s+\S+\s+(DISABLE|ENABLE) KEYS\s*;?`)
	rawSQL = reAlter.ReplaceAllStringFunc(rawSQL, func(s string) string {
		return "-- " + s
	})

	// 分割语句
	reSplit := regexp.MustCompile(`;[\s]*\n`)
	statements := reSplit.Split(rawSQL, -1)

	return statements, nil
}
