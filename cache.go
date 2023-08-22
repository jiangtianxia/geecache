package geecache

import (
	"geecache/lru"
	"sync"
	"time"
)

type cache struct {
	mu       sync.Mutex
	lru      *lru.LRUCache
	capacity int64
}

// 定时清除过期key的任务协程
func (c *cache) startExpiryCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.lru.CleanupExpiredKeys()
			c.mu.Unlock()
		}
	}
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()

	if c.lru == nil {
		c.lru = lru.NewCache(c.capacity, nil)
		go c.startExpiryCleanup(10 * time.Minute)
	}
	c.lru.Add(key, value)
	c.mu.Unlock()
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}

func (c *cache) expire(key string, second int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	c.lru.Expire(key, second)
}
