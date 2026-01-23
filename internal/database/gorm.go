package database

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewGORM(dbPath string) (*gorm.DB, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SQLite 在 Windows 上有并发限制，限制为单连接
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// NewGormLogger 创建 GORM 日志适配器
func NewGormLogger() gormlogger.Interface {
	return &gormLogger{}
}

type gormLogger struct{}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Info(msg, data)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Warn(msg, data)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.Error(msg, data)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()

	if err != nil {
		logger.Errorf("SQL error: %s, duration: %v, error: %v", sql, elapsed, err)
	} else if elapsed > 200*time.Millisecond {
		logger.Warnf("Slow SQL: %s, duration: %v", sql, elapsed)
	} else {
		logger.Debugf("SQL: %s, duration: %v", sql, elapsed)
	}
}

// AutoMigrate 自动迁移数据库模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Program{},
		&models.Version{},
		&models.Token{},
		&models.EncryptionKey{},
	)
}
