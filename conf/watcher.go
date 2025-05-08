package conf

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Watcher 文件监听器
type Watcher struct {
	mu       sync.Mutex
	path     string
	modTime  time.Time
	interval time.Duration
	done     chan struct{}
	onChange func()
}

// NewWatcher 创建文件监听器
func NewWatcher(path string, interval time.Duration, onChange func()) *Watcher {
	return &Watcher{
		path:     path,
		interval: interval,
		done:     make(chan struct{}),
		onChange: onChange,
	}
}

// Start 启动监听
func (w *Watcher) Start() error {
	// 获取初始文件信息
	info, err := os.Stat(w.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("get file info failed: %v", err)
		}
		// 文件不存在时记录当前时间
		w.modTime = time.Now()
	} else {
		w.modTime = info.ModTime()
	}

	// 启动监听协程
	go w.watch()
	return nil
}

// watch 监听文件变化
func (w *Watcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkFile()
		case <-w.done:
			return
		}
	}
}

// checkFile 检查文件是否变化
func (w *Watcher) checkFile() {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, err := os.Stat(w.path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("get file info failed: %v\n", err)
		}
		return
	}

	// 检查修改时间是否变化
	if info.ModTime().After(w.modTime) {
		w.modTime = info.ModTime()
		if w.onChange != nil {
			w.onChange()
		}
	}
}

// Stop 停止监听
func (w *Watcher) Stop() {
	close(w.done)
}
