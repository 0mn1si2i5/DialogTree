// Path: ./core/init_db.go

package core

import (
	"dialogTree/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
)

func InitDB() (db *gorm.DB) {
	gdb := global.Config.DB

	db, err := gorm.Open(postgres.Open(gdb.DSN()), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			db = createDB()
		} else {
			logrus.Fatalln("DB open error: ", err)
		}
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 20)

	logrus.Infof("DataBase [%s:%d] connection successful", gdb.Host, gdb.Port)
	return
}

func createDB() *gorm.DB {
	gdb := global.Config.DB
	db, err := gorm.Open(postgres.Open(gdb.DSNWithoutDB()), &gorm.Config{})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	dbName := gdb.DBname
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s WITH ENCODING 'UTF8';", dbName)

	if err := db.Exec(createDBSQL).Error; err != nil {
		logrus.Fatalln("Create database error: ", err.Error())
	}

	logrus.Infoln("Database created: ", dbName)

	// ✅ PostgreSQL 不支持 USE，要重新连接
	db, err = gorm.Open(postgres.Open(gdb.DSN()), &gorm.Config{})
	if err != nil {
		logrus.Fatalln("Reconnect error: ", err)
	}

	return db
}
