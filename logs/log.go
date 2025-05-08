package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Log struct {
	mu         sync.Mutex
	out        io.Writer
	file       *os.File
	level      Level
	filePath   string
	formatter  Formatter
	forceColor bool
}

// NewLogger 创建日志记录器
func NewLogger(config *Config) (*Log, error) {
	l := &Log{
		level:      InfoLevel,
		out:        io.Discard,
		formatter:  NewTextFormatter(),
		forceColor: config.ForceColor,
	}

	if config != nil {
		l.level = config.Level
		if config.Formatter != nil {
			l.formatter = config.Formatter
		}
		var writers []io.Writer
		if config.Console {
			writers = append(writers, os.Stdout)
		}
		if config.FilePath != "" {
			if err := l.SetOutput(config.FilePath); err != nil {
				return nil, err
			}
			if l.file != nil {
				writers = append(writers, l.file)
			}
		}
		if len(writers) > 0 {
			l.out = io.MultiWriter(writers...)
		}
	}

	return l, nil
}

// SetOutput 设置日志输出
func (l *Log) SetOutput(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 创建日志目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 打开日志文件
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// 关闭旧文件
	if l.file != nil {
		l.file.Close()
	}

	l.file = f
	l.filePath = path
	return nil
}

// 获取日志级别对应的颜色
func (l *Log) levelColor(level Level) string {
	if !l.forceColor {
		return ""
	}

	switch level {
	case DebugLevel:
		return colorGray
	case InfoLevel:
		return colorGreen
	case WarnLevel:
		return colorYellow
	case ErrorLevel:
		return colorRed
	case FatalLevel:
		return colorPurple
	default:
		return colorReset
	}
}

// 格式化日志消息
func (l *Log) format(level Level, msg string) string {
	color := l.levelColor(level)
	reset := ""
	if color != "" {
		reset = colorReset
	}

	return fmt.Sprintf("%s %s%s",
		color,
		msg,
		reset,
	)
}

// log 内部日志方法
func (l *Log) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var file, funcName string
	var line int

	for i := 1; i < 15; i++ {
		pc, f, l, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if !strings.Contains(f, "pkg/logs") {
			file = f
			line = l
			if fn := runtime.FuncForPC(pc); fn != nil {
				funcName = fn.Name()
			}
			break
		}
	}

	fmt.Fprint(l.out, l.format(level, l.formatter.Format(level, fmt.Sprintf(format, args...), &CallerInfo{
		File:     file,
		Line:     line,
		Function: funcName,
	})))

	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug 级别日志
func (l *Log) Debug(args ...interface{}) {
	l.log(DebugLevel, "%s", fmt.Sprint(args...))
}

func (l *Log) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info 级别日志
func (l *Log) Info(args ...interface{}) {
	l.log(InfoLevel, "%s", fmt.Sprint(args...))
}

func (l *Log) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn 级别日志
func (l *Log) Warn(args ...interface{}) {
	l.log(WarnLevel, "%s", fmt.Sprint(args...))
}

func (l *Log) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error 级别日志
func (l *Log) Error(args ...interface{}) {
	l.log(ErrorLevel, "%s", fmt.Sprint(args...))
}

func (l *Log) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal 级别日志
func (l *Log) Fatal(args ...interface{}) {
	l.log(FatalLevel, "%s", fmt.Sprint(args...))
}

func (l *Log) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

func (l *Log) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Log) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Close 关闭日志
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
