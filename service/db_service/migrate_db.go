// Path: ./service/db_service/migrate_db.go

package db_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"github.com/sirupsen/logrus"
)

func MigrateDB() {
	// 表迁移
	err := global.DB.AutoMigrate(
		&models.CategoryModel{},
		&models.SessionModel{},
		&models.DialogModel{},
		&models.ConversationModel{},
		&models.ImageModel{},
	)
	if err != nil {
		logrus.Errorf("failed to migrate DB: %s\n", err)
		return
	}
	logrus.Info("DB migration successful")

	// 检查并创建默认分类
	createDefaultCategory()
}

func createDefaultCategory() {
	var count int64
	global.DB.Model(&models.CategoryModel{}).Count(&count)
	
	if count == 0 {
		defaultCategory := models.CategoryModel{
			Name: "General",
		}
		
		if err := global.DB.Create(&defaultCategory).Error; err != nil {
			logrus.Errorf("failed to create default category: %s\n", err)
			return
		}
		
		logrus.Info("Default category 'General' created successfully")
	} else {
		logrus.Infof("Categories already exist (%d found), skipping default category creation", count)
	}
}
