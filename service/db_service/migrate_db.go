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
}
