// Path: ./flags/flag_db.go

package flags

import (
	"dialogTree/global"
	"dialogTree/models"
	"github.com/sirupsen/logrus"
)

func FlagDB() {
	// 表迁移
	err := global.DB.AutoMigrate(
		&models.CategoryModel{},
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
