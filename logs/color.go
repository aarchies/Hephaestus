package logs

// ANSI 颜色代码
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Red        = "\033[31m"
	Green      = "\033[32m"
	Yellow     = "\033[33m"
	Blue       = "\033[34m"
	Purple     = "\033[35m"
	Cyan       = "\033[36m"
	White      = "\033[37m"
	BoldRed    = "\033[1;31m"
	BoldGreen  = "\033[1;32m"
	BoldYellow = "\033[1;33m"
	BoldBlue   = "\033[1;34m"
	BoldPurple = "\033[1;35m"
	BoldCyan   = "\033[1;36m"
	BoldWhite  = "\033[1;37m"
)

// 获取日志级别对应的颜色
func getLevelColor(level Level) string {
	switch level {
	case DebugLevel:
		return Purple
	case InfoLevel:
		return Green
	case WarnLevel:
		return Yellow
	case ErrorLevel:
		return Red
	case FatalLevel:
		return BoldRed
	default:
		return Reset
	}
}

// colorize 给文本添加颜色
func colorize(color string, text string) string {
	return color + text + Reset
}
