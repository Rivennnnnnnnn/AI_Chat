package db

import (
	"AI_Chat/pkg/utils"
	"fmt"
	"time"

	"gorm.io/driver/mysql" // 如果用 postgres 就换成 gorm.io/driver/postgres
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

/*
mysql配置结构体

	type MysqlConfig struct {
		Host         string
		Port         int
		User         string
		Password     string
		DBName       string
		MaxIdleConns int
		MaxOpenConns int
	}
*/
func InitDB(cfg utils.MysqlConfig) (*gorm.DB, error) {
	// 实际开发中，这些参数应该从配置文件读取
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	newLogger := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger, // 设置日志
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 设置连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	return db, nil
}
