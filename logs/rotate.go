package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Rotator 日志轮转器
type Rotator struct {
	config *Config
	log    *Log
}

// NewRotator 创建日志轮转器
func NewRotator(config *Config, log *Log) *Rotator {
	return &Rotator{
		config: config,
		log:    log,
	}
}

// Rotate 执行日志轮转
func (r *Rotator) Rotate() error {
	if r.log.file == nil {
		return nil
	}

	// 检查文件大小
	info, err := r.log.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < r.config.MaxSize*1024*1024 {
		return nil
	}

	// 关闭当前文件
	r.log.file.Close()

	// 生成新文件名
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", r.log.filePath, timestamp)

	// 重命名文件
	if err := os.Rename(r.log.filePath, newPath); err != nil {
		return err
	}

	// 打开新文件
	return r.log.SetOutput(r.log.filePath)
}

// Clean 清理旧日志
func (r *Rotator) Clean() error {
	dir := filepath.Dir(r.log.filePath)
	pattern := filepath.Base(r.log.filePath) + ".*"

	// 获取所有日志文件
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return err
	}

	// 检查文件时间
	now := time.Now()
	maxAge := time.Duration(r.config.MaxAge) * 24 * time.Hour

	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			os.Remove(path)
		}
	}

	return nil
}
