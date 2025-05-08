package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Manager 配置管理器
type Manager[T any] struct {
	mu       sync.RWMutex
	path     string
	config   *T
	onChange func(*T)
	watcher  *Watcher
}

// New 创建配置管理器
func New[T any](path string, onChange func(*T)) (*Manager[T], error) {
	m := &Manager[T]{
		path:     path,
		config:   new(T),
		onChange: onChange,
	}

	// 首次加载配置
	if err := m.Load(); err != nil {
		return nil, err
	}

	// 创建并启动监听器
	m.watcher = NewWatcher(path, time.Second, func() {
		if err := m.Load(); err != nil {
			fmt.Printf("reload config failed: %v\n", err)
			return
		}
		if m.onChange != nil {
			m.onChange(m.config)
		}
	})

	if err := m.watcher.Start(); err != nil {
		return nil, fmt.Errorf("start watcher failed: %v", err)
	}

	return m, nil
}

// Load 加载配置
func (m *Manager[T]) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("read config file failed: %v", err)
	}

	// 解析配置
	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("parse config failed: %v", err)
	}

	// 加载环境变量
	if err := LoadEnv(m.config); err != nil {
		return fmt.Errorf("load env failed: %v", err)
	}

	return nil
}

// Get 获取配置
func (m *Manager[T]) Get() *T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Close 关闭配置管理器
func (m *Manager[T]) Close() error {
	m.watcher.Stop()
	return nil
}
