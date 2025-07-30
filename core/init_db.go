// Path: ./core/init_db.go

package core

import (
	"dialogTree/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"time"
)

func InitDB() (db *gorm.DB) {
	gdb := global.Config.DB

	var dialector gorm.Dialector
	switch gdb.Source {
	case "mysql":
		dialector = mysql.Open(gdb.DSN())
	case "pgsql", "postgres":
		dialector = postgres.Open(gdb.DSN())
	case "sqlite":
		dialector = sqlite.Open(gdb.DSN())
	default:
		logrus.Fatalln("Unsupported database source: ", gdb.Source)
	}

	var gormConfig gorm.Config
	if gdb.Source == "mysql" {
		gormConfig = gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}
	} else {
		gormConfig = gorm.Config{}
	}

	db, err := gorm.Open(dialector, &gormConfig)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") && gdb.Source != "sqlite" {
			db = createDB()
		} else {
			logrus.Fatalln("DB open error: ", err)
		}
	}

	if gdb.Source != "sqlite" {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Minute * 20)
		logrus.Infof("DataBase [%s:%d] connection successful", gdb.Host, gdb.Port)
	} else {
		logrus.Infof("SQLite database [%s] connection successful", gdb.DBname)
	}
	
	return
}

func createDB() *gorm.DB {
	gdb := global.Config.DB
	
	var dialector gorm.Dialector
	switch gdb.Source {
	case "mysql":
		dialector = mysql.Open(gdb.DSNWithoutDB())
	case "pgsql", "postgres":
		dialector = postgres.Open(gdb.DSNWithoutDB())
	default:
		logrus.Fatalln("Cannot create database for source: ", gdb.Source)
	}

	var gormConfig gorm.Config
	if gdb.Source == "mysql" {
		gormConfig = gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}
	} else {
		gormConfig = gorm.Config{}
	}

	db, err := gorm.Open(dialector, &gormConfig)
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	dbName := gdb.DBname
	var createDBSQL string
	switch gdb.Source {
	case "mysql":
		createDBSQL = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4;", dbName)
	case "pgsql", "postgres":
		createDBSQL = fmt.Sprintf("CREATE DATABASE %s WITH ENCODING 'UTF8';", dbName)
	}

	if err := db.Exec(createDBSQL).Error; err != nil {
		logrus.Fatalln("Create database error: ", err.Error())
	}

	logrus.Infoln("Database created: ", dbName)

	// 重新连接到新创建的数据库
	switch gdb.Source {
	case "mysql":
		dialector = mysql.Open(gdb.DSN())
	case "pgsql", "postgres":
		dialector = postgres.Open(gdb.DSN())
	}
	
	db, err = gorm.Open(dialector, &gormConfig)
	if err != nil {
		logrus.Fatalln("Reconnect error: ", err)
	}

	return db
}
