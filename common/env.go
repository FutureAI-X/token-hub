package common

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var envOnce sync.Once

// ensureEnvLoaded 加载 .env 环境变量（幂等）。
// 必须在读取任何密钥前调用——package init 早于 main 的 godotenv.Load()，
// 因此密钥获取应改为延迟调用本函数，确保 .env 已先被加载。
func ensureEnvLoaded() {
	envOnce.Do(func() {
		_ = godotenv.Load()
	})
}

// GetEnvOrDefault 获取环境变量，如果不存在则返回默认值
func GetEnvOrDefault(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %d", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

// GetEnvOrDefaultString 获取字符串环境变量
func GetEnvOrDefaultString(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}

// GetEnvOrDefaultBool 获取布尔环境变量
func GetEnvOrDefaultBool(env string, defaultValue bool) bool {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %t", env, err.Error(), defaultValue))
		return defaultValue
	}
	return b
}
