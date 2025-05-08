package logs

import (
	"encoding/json"
	"fmt"
	"time"
)

// Formatter 日志格式化器接口
type Formatter interface {
	Format(level Level, msg string, caller *CallerInfo) string
}

// CallerInfo 调用者信息
type CallerInfo struct {
	File     string
	Line     int
	Function string
}

// TextFormatter 文本格式化器
type TextFormatter struct {
	// 时间格式
	TimeFormat string
	// 是否显示完整路径
	FullPath   bool
	ForceColor bool // 强制启用颜色
	NoColor    bool // 禁用颜色
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		TimeFormat: "2006-01-02 15:04:05.000",
		FullPath:   true,
	}
}

func (f *TextFormatter) Format(level Level, msg string, caller *CallerInfo) string {
	timestamp := time.Now().Format(f.TimeFormat)
	callerInfo := formatCaller(caller, f.FullPath)

	levelText := levelString(level)
	if f.ForceColor {
		levelText = colorize(getLevelColor(level), levelText)
	}

	return fmt.Sprintf("[%s] [%s] %s %s\n", levelText, timestamp, callerInfo, msg)
}

// JSONFormatter JSON格式化器
type JSONFormatter struct {
	TimeFormat string
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		TimeFormat: time.RFC3339,
	}
}

func (f *JSONFormatter) Format(level Level, msg string, caller *CallerInfo) string {
	data := make(map[string]interface{})
	data["time"] = time.Now().Format(f.TimeFormat)
	data["level"] = levelString(level)
	data["msg"] = msg
	data["file"] = caller.File
	data["line"] = caller.Line

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error marshaling log entry: %v", err)
	}

	return string(bytes) + "\n"
}
