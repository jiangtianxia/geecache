package lru

import (
	"container/list"
	"math/rand"
	"time"
)

type LRUCache struct {
	capacity   int64                         // 最大缓存数量
	length     int64                         // 当前缓存数量
	list       *list.List                    // 数据链表
	cache      map[string]*list.Element      // 缓存map
	expireDict map[string]int64              // 过期字典
	OnEvivted  func(key string, value Value) // 某条记录被移除时的回调函数, 可以为nil
}

// 节点
type entry struct {
	key   string
	value Value
}

// 任意类型
type Value interface {
}

func NewCache(capacity int64, onEvicted func(string, Value)) *LRUCache {
	lru := &LRUCache{
		capacity:   capacity,
		length:     0,
		list:       list.New(),
		cache:      make(map[string]*list.Element),
		expireDict: make(map[string]int64),
		OnEvivted:  onEvicted,
	}

	return lru
}

// 查找功能
func (c *LRUCache) Get(key string) (Value, bool) {
	// 1、从字典中找到对应的双向链表的结点
	// 2、判断该结点是否已经过期, 过期则删除
	// 3、不过期, 将该结点移动到队首
	if ele, ok := c.cache[key]; ok {
		if c.CheckKey(key) {
			c.RemoveNode(ele)
			return nil, false
		}

		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// 删除功能
func (c *LRUCache) RemoveOldest() {
	ele := c.list.Back()
	if ele != nil {
		c.RemoveNode(ele)
		kv := ele.Value.(*entry)
		if c.OnEvivted != nil {
			c.OnEvivted(kv.key, kv.value)
		}
	}
}

// 新增/修改功能
func (c *LRUCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 如果键存在, 则先判断是否过期, 更新对应节点的值, 并将该节点移动到队尾
		if c.CheckKey(key) {
			c.RemoveNode(ele)
			ele := c.list.PushFront(&entry{key, value})
			c.cache[key] = ele
			c.length++
			return
		}
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.value = value
		return
	}
	// 不存在则新增
	ele := c.list.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.length++

	// 如果缓存数量超过限制, 则移除最少访问的节点
	for c.capacity != 0 && c.capacity < c.length {
		c.RemoveOldest()
	}
}

// 设置过期时间
func (c *LRUCache) Expire(key string, second int64) {
	c.expireDict[key] = time.Now().Add(time.Duration(second) * time.Second).Unix()
}

// 检测key是否已经过期
func (c *LRUCache) CheckKey(key string) bool {
	if t, ok := c.expireDict[key]; ok && t < time.Now().Unix() {
		return true
	}
	return false
}

// 定期清理key
func (c *LRUCache) CleanupExpiredKeys() {
	checkKeysIndex := make(map[int64]struct{}, 5)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 5; i++ {
		keyIdx := rand.Int63n(c.length)
		if _, ok := checkKeysIndex[keyIdx]; ok {
			i--
		} else {
			checkKeysIndex[keyIdx] = struct{}{}
		}
	}

	now := time.Now().Unix()
	var i int64
	for key, expireTime := range c.expireDict {
		if _, ok := checkKeysIndex[i]; ok && expireTime <= now {
			c.RemoveNode(c.cache[key])
		}
		i++
	}
}

// 删除结点
func (c *LRUCache) RemoveNode(node *list.Element) {
	c.list.Remove(node)
	c.length--
	kv := node.Value.(*entry)
	delete(c.cache, kv.key)
	delete(c.expireDict, kv.key)
}
