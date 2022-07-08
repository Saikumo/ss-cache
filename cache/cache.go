package cache

import (
	"saikumo.org/cache/lru"
	"sync"
)

// 缓存结构体
type cache struct {
	mu            sync.Mutex
	lru           *lru.Cache
	cacheMaxBytes uint64
}

//缓存添加
func (c *cache) add(key string, value ByteView) {
	//加锁
	c.mu.Lock()
	defer c.mu.Unlock()
	//懒加载
	if c.lru == nil {
		c.lru = lru.NewCache(c.cacheMaxBytes, nil)
	}
	c.lru.Add(key, value)
}

//缓存获取
func (c *cache) get(key string) (value ByteView, ok bool) {
	//加锁
	c.mu.Lock()
	defer c.mu.Unlock()
	//没有缓存
	if c.lru == nil {
		return
	}

	if v, exist := c.lru.Get(key); exist {
		return v.(ByteView), exist
	}
	return
}
