// Path: ./service/demo_service/enter.go

package demo_service

import (
	"dialogTree/global"
	"github.com/sirupsen/logrus"
	"time"
)

func Demo() {
	if global.DemoStart {
		return
	}
	global.DemoStart = true
	hours := time.Duration(global.Config.System.DemoTimer)
	ticker := time.NewTicker(hours * time.Hour)
	go func() {
		logrus.Info("Demo mode started, DB will be reset in ", hours, " hours.")
		for range ticker.C {
			initDB()
		}
	}()
}

// 初始化数据库
func initDB() {
	// 清空所有表 → 执行sample_data.sql

	global.DemoStart = false
}
