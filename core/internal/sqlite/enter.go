package sqlite

import (
	"fmt"

	"github.com/baibanzz/jdk/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqlite3(sqlite3 model.Sqlite3, loggers logger.Interface) (*gorm.DB, error) {
	dsn := sqlite3.Dsn()
	if loggers == nil {
		loggers = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: loggers,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池（SQLite 建议使用默认配置）
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return db, nil
}
