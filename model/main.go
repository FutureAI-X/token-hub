package model

import (
	"fmt"
	"os"
	"time"

	"github.com/FutureAI/token-hub/common"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// newGormConfig 创建 GORM 配置
func newGormConfig() *gorm.Config {
	return &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	}
}

// openPostgreSQL 打开 PostgreSQL 连接
func openPostgreSQL(dsn string) (*gorm.DB, error) {
	common.SysLog("using PostgreSQL as database")
	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), newGormConfig())
}

// InitDB 初始化数据库
// 后续扩展 MySQL: 在此函数中根据 SQL_DSN 前缀选择驱动
func InitDB() error {
	dsn := os.Getenv("SQL_DSN")
	if dsn == "" {
		return fmt.Errorf("SQL_DSN environment variable is required")
	}

	db, err := openPostgreSQL(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if common.DebugEnabled {
		db = db.Debug()
	}

	DB = db

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 10))
	sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 100))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 60)))

	// 数据库迁移
	common.SysLog("database migration started")
	err = migrateDB()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 创建默认模型数据
	err = createDefaultModels()
	if err != nil {
		common.SysError("failed to create default models: " + err.Error())
	}

	return nil
}

// migrateDB 数据库迁移
func migrateDB() error {
	return DB.AutoMigrate(
		&User{},
		&Token{},
		&Model{},
	)
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
