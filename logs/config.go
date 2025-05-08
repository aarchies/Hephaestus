package logs

// Config 日志配置
type Config struct {
	Level      Level     `json:"level"`       // 日志级别
	FilePath   string    `json:"file_path"`   // 日志文件路径
	MaxSize    int64     `json:"max_size"`    // 单个日志文件最大大小(MB)
	MaxAge     int       `json:"max_age"`     // 日志保留天数
	Compress   bool      `json:"compress"`    // 是否压缩
	Formatter  Formatter `json:"formatter"`   // 日志格式化器
	ForceColor bool      `json:"force_color"` // 强制启用颜色
	NoColor    bool      `json:"no_color"`    // 禁用颜色
	Console    bool      `json:"console"`     // 是否输出到控制台
}
