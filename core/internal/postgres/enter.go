package postgres

import (
	"fmt"
	"time"

	"github.com/baibanzz/jdk/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgreSql(postgreSql model.PostgreSql, loggers logger.Interface) (*gorm.DB, error) {
	dsn := postgreSql.Dsn()
	if loggers == nil {
		loggers = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: loggers,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
