// Path: ./core/init_db.go

package core

import (
	"dialogTree/global"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func InitDB() (db *gorm.DB) {
	gdb := global.Config.DB

	db, err := gorm.Open(mysql.Open(gdb.DSN()), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") {
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
	// 创建数据库
	gdb := global.Config.DB
	db, err := gorm.Open(mysql.Open(gdb.DSNWithoutDB()), &gorm.Config{})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	dbName := global.Config.DB.DBname
	createDBSQL := "CREATE DATABASE IF NOT EXISTS " + dbName + " DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;"
	if err := db.Exec(createDBSQL).Error; err != nil {
		logrus.Fatalln("Create database error: ", err.Error())
		return nil
	}
	logrus.Infoln("Database created: ", dbName)
	db.Exec("USE " + dbName + ";")
	return db
}
