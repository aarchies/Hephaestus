package logs

import (
	"os"
	"sync"
)

var (
	globalLogger     *Log
	globalLoggerOnce sync.Once
)

// GetLogger 获取全局日志实例
func GetLogger() *Log {
	globalLoggerOnce.Do(func() {
		if globalLogger == nil {
			globalLogger = &Log{
				level:     InfoLevel,
				out:       os.Stdout,
				formatter: NewTextFormatter(),
			}
		}
	})
	return globalLogger
}

// SetLogger 设置全局日志实例
func SetLogger(logger *Log) {
	globalLoggerOnce.Do(func() {}) // 确保 Once 已经执行过
	globalLogger = logger
}

// Debug 全局 Debug 日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 全局 Info 日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 全局 Warn 日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 全局 Error 日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 全局 Fatal 日志
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
