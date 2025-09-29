package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryItem 内存存储项
type MemoryItem struct {
	Value      interface{}
	ExpireTime time.Time
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	data sync.Map
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{}

	// 启动清理过期数据的goroutine
	go storage.cleanupExpired()

	return storage
}

// Set 设置键值
func (m *MemoryStorage) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	var expireTime time.Time
	if expire > 0 {
		expireTime = time.Now().Add(expire)
	}

	item := &MemoryItem{
		Value:      value,
		ExpireTime: expireTime,
	}

	m.data.Store(key, item)
	return nil
}

// Get 获取值
func (m *MemoryStorage) Get(ctx context.Context, key string) (interface{}, error) {
	value, ok := m.data.Load(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	item := value.(*MemoryItem)

	// 检查是否过期
	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		m.data.Delete(key)
		return nil, fmt.Errorf("key expired: %s", key)
	}

	return item.Value, nil
}

// Delete 删除键
func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.data.Delete(key)
	return nil
}

// Exists 检查键是否存在
func (m *MemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.Get(ctx, key)
	return err == nil, nil
}

// Keys 获取匹配的键列表
func (m *MemoryStorage) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string

	m.data.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if m.matchPattern(keyStr, pattern) {
			keys = append(keys, keyStr)
		}
		return true
	})

	return keys, nil
}

// matchPattern 简单的模式匹配
func (m *MemoryStorage) matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// 简化的通配符匹配
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(key, prefix)
	}

	return key == pattern
}

// cleanupExpired 清理过期数据
func (m *MemoryStorage) cleanupExpired() {
	defer func() {
		if r := recover(); r != nil {
			// 记录panic信息，但不让进程退出
			fmt.Printf("Memory storage cleanup panic recovered: %v\n", r)
			// 重新启动清理goroutine
			go m.cleanupExpired()
		}
	}()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.data.Range(func(key, value interface{}) bool {
			item := value.(*MemoryItem)
			if !item.ExpireTime.IsZero() && now.After(item.ExpireTime) {
				m.data.Delete(key)
			}
			return true
		})
	}
}
