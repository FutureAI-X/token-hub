package common

// DatabaseType 数据库类型
// 后续扩展 MySQL: 添加 DatabaseTypeMySQL DatabaseType = "mysql"
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgres"
)

var mainDatabaseType = DatabaseTypePostgreSQL

// MainDatabaseType 获取主数据库类型
func MainDatabaseType() DatabaseType {
	return mainDatabaseType
}

// SetMainDatabaseType 设置主数据库类型
func SetMainDatabaseType(databaseType DatabaseType) {
	mainDatabaseType = databaseType
}

// UsingMainDatabase 检查是否使用指定的主数据库类型
func UsingMainDatabase(databaseType DatabaseType) bool {
	return mainDatabaseType == databaseType
}
