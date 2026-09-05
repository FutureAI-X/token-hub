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

	// 更新表注释
	err = addTableComments()
	if err != nil {
		common.SysError("failed to add table comments: " + err.Error())
	}

	// 创建默认 root 用户（如果不存在）
	err = CreateRootUserIfNeed()
	if err != nil {
		common.SysError("failed to create root user: " + err.Error())
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
	err := DB.AutoMigrate(
		&User{},
		&Token{},
		&Vendor{},
		&Model{},
		&VendorModel{},
		&Endpoint{},
		&ModelEndpoint{},
		&Task{},
		&QuotaRule{},
		&QuotaRuleItem{},
		&QuotaLog{},
	)
	if err != nil {
		return err
	}

	// 添加表和字段注释
	return addTableComments()
}

// addTableComments 为 PostgreSQL 表和字段添加注释
func addTableComments() error {
	comments := []string{
		// users 表注释
		`COMMENT ON TABLE users IS '用户表，存储系统用户信息'`,
		`COMMENT ON COLUMN users.id IS '用户唯一标识，自增主键'`,
		`COMMENT ON COLUMN users.username IS '用户名，用于登录，全局唯一'`,
		`COMMENT ON COLUMN users.password IS '密码哈希值，使用 bcrypt 加密存储'`,
		`COMMENT ON COLUMN users.display_name IS '显示名称，用于界面展示'`,
		`COMMENT ON COLUMN users.role IS '用户角色：1=普通用户, 10=管理员, 100=root'`,
		`COMMENT ON COLUMN users.status IS '用户状态：1=启用, 2=禁用'`,
		`COMMENT ON COLUMN users.email IS '用户邮箱，用于通知和找回密码'`,
		`COMMENT ON COLUMN users.credits IS '用户当前积分'`,
		`COMMENT ON COLUMN users.used_credits IS '已使用积分'`,
		`COMMENT ON COLUMN users.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN users.updated_at IS '记录最后更新时间'`,
		`COMMENT ON COLUMN users.deleted_at IS '软删除时间戳，非空表示已删除'`,

		// tokens 表注释
		`COMMENT ON TABLE tokens IS 'API Token 表，存储用户 API 访问令牌'`,
		`COMMENT ON COLUMN tokens.id IS 'Token 唯一标识，自增主键'`,
		`COMMENT ON COLUMN tokens.user_id IS '所属用户ID，关联 users 表'`,
		`COMMENT ON COLUMN tokens.key IS 'Token 密钥，用于 API 认证'`,
		`COMMENT ON COLUMN tokens.name IS 'Token 名称，便于用户识别'`,
		`COMMENT ON COLUMN tokens.status IS 'Token 状态：1=启用, 2=禁用'`,
		`COMMENT ON COLUMN tokens.expired_time IS '过期时间戳，-1 表示永不过期'`,
		`COMMENT ON COLUMN tokens.remain_quota IS '剩余配额，-1 表示无限制'`,
		`COMMENT ON COLUMN tokens.used_quota IS '已使用配额'`,
		`COMMENT ON COLUMN tokens.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN tokens.updated_at IS '记录最后更新时间'`,
		`COMMENT ON COLUMN tokens.deleted_at IS '软删除时间戳'`,

		// vendors 表注释
		`COMMENT ON TABLE vendors IS '供应商表，存储 AI 模型供应商信息'`,
		`COMMENT ON COLUMN vendors.id IS '供应商唯一标识，自增主键'`,
		`COMMENT ON COLUMN vendors.name IS '供应商名称，全局唯一'`,
		`COMMENT ON COLUMN vendors.description IS '供应商描述'`,
		`COMMENT ON COLUMN vendors.icon IS '供应商图标标识'`,
		`COMMENT ON COLUMN vendors.status IS '供应商状态：1=启用, 2=禁用'`,
		`COMMENT ON COLUMN vendors.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN vendors.updated_at IS '记录最后更新时间'`,
		`COMMENT ON COLUMN vendors.deleted_at IS '软删除时间戳'`,

		// models 表注释
		`COMMENT ON TABLE models IS '模型表，存储可用的 AI 模型信息'`,
		`COMMENT ON COLUMN models.id IS '模型唯一标识，自增主键'`,
		`COMMENT ON COLUMN models.name IS '模型名称，全局唯一，用于 API 调用'`,
		`COMMENT ON COLUMN models.description IS '模型描述信息'`,
		`COMMENT ON COLUMN models.tags IS '模型标签，逗号分隔'`,
		`COMMENT ON COLUMN models.owner IS '模型所有者/提供商'`,
		`COMMENT ON COLUMN models.status IS '模型状态：1=启用, 2=禁用'`,
		`COMMENT ON COLUMN models.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN models.updated_at IS '记录最后更新时间'`,
		`COMMENT ON COLUMN models.deleted_at IS '软删除时间戳'`,

		// quota_rules 表注释
		`COMMENT ON TABLE quota_rules IS '积分扣除规则表，存储模型的积分扣除算法'`,
		`COMMENT ON COLUMN quota_rules.id IS '规则唯一标识，自增主键'`,
		`COMMENT ON COLUMN quota_rules.model_id IS '关联的模型ID，关联 models 表，唯一'`,
		`COMMENT ON COLUMN quota_rules.rule_type IS '规则类型：per_request=按次计费'`,
		`COMMENT ON COLUMN quota_rules.base_price IS '基础积分价格（每次请求扣除的积分数量）'`,
		`COMMENT ON COLUMN quota_rules.description IS '规则描述'`,
		`COMMENT ON COLUMN quota_rules.status IS '规则状态：1=启用, 2=禁用'`,
		`COMMENT ON COLUMN quota_rules.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN quota_rules.updated_at IS '记录最后更新时间'`,
		`COMMENT ON COLUMN quota_rules.deleted_at IS '软删除时间戳，非空表示已删除'`,

		// quota_rule_items 表注释
		`COMMENT ON TABLE quota_rule_items IS '积分规则参数映射表，存储差异化计费的参数配置'`,
		`COMMENT ON COLUMN quota_rule_items.id IS '项唯一标识，自增主键'`,
		`COMMENT ON COLUMN quota_rule_items.rule_id IS '关联的规则ID，关联 quota_rules 表'`,
		`COMMENT ON COLUMN quota_rule_items.param_path IS '请求参数路径（如 size, quality, model）'`,
		`COMMENT ON COLUMN quota_rule_items.param_value IS '参数值（如 1024x1024, high, gpt-4）'`,
		`COMMENT ON COLUMN quota_rule_items.price IS '该参数值对应的积分价格'`,
		`COMMENT ON COLUMN quota_rule_items.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN quota_rule_items.updated_at IS '记录最后更新时间'`,

		// quota_logs 表注释
		`COMMENT ON TABLE quota_logs IS '积分日志表，记录积分扣除和退还'`,
		`COMMENT ON COLUMN quota_logs.id IS '日志唯一标识，自增主键'`,
		`COMMENT ON COLUMN quota_logs.user_id IS '用户ID，关联 users 表'`,
		`COMMENT ON COLUMN quota_logs.task_id IS '关联的任务ID'`,
		`COMMENT ON COLUMN quota_logs.amount IS '积分数量（正数）'`,
		`COMMENT ON COLUMN quota_logs.type IS '操作类型：deduct=扣除, refund=退还'`,
		`COMMENT ON COLUMN quota_logs.remark IS '备注说明'`,
		`COMMENT ON COLUMN quota_logs.created_at IS '记录创建时间'`,
		`COMMENT ON COLUMN quota_logs.deleted_at IS '软删除时间戳'`,
	}

	for _, comment := range comments {
		if err := DB.Exec(comment).Error; err != nil {
			// 注释失败不影响正常使用，只记录警告
			common.SysError("failed to add comment: " + err.Error())
		}
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
