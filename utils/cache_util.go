package utils

import (
	"sync"
	"time"
)

// CacheItem 表示缓存中的一个项目
type CacheItem struct {
	Value     interface{}
	ExpiredAt time.Time
}

// Cache 通用缓存结构
type Cache struct {
	mutex sync.RWMutex
	items map[string]CacheItem
}

// NewCache 创建一个新的缓存实例
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

// Set 设置缓存，带有过期时间
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Value:     value,
		ExpiredAt: time.Now().Add(duration),
	}
}

// Get 获取缓存值，如果不存在或已过期则返回nil和false
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	item, exists := c.items[key]
	c.mutex.RUnlock()

	// 检查是否存在
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiredAt) {
		// 异步删除过期项
		go func() {
			c.Delete(key)
		}()
		return nil, false
	}

	return item.Value, true
}

// Delete 从缓存中删除一个项目
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// 全局缓存实例
var GlobalCache = NewCache()
