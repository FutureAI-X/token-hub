package common

import (
	"fmt"
	"log"
	"os"
)

// DebugEnabled 是否启用调试模式
var DebugEnabled = false

// SysLog 系统日志
func SysLog(msg string) {
	log.Printf("[TOKEN-HUB] %s", msg)
}

// SysError 系统错误日志
func SysError(msg string) {
	log.Printf("[TOKEN-HUB] [ERROR] %s", msg)
}

// FatalLog 致命错误日志
func FatalLog(msg string) {
	log.Fatalf("[TOKEN-HUB] [FATAL] %s", msg)
	os.Exit(1)
}

// SysLogf 格式化系统日志
func SysLogf(format string, args ...interface{}) {
	SysLog(fmt.Sprintf(format, args...))
}

// SysErrorf 格式化系统错误日志
func SysErrorf(format string, args ...interface{}) {
	SysError(fmt.Sprintf(format, args...))
}
