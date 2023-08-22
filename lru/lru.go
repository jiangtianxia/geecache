package lru

import "container/list"

type Cache struct {
	maxBytes int64 // 最大内存
	nbytes   int64 // 已使用内存
	l1       *list.List
	cache    map[string]*list.Element

	OnEvivted func(key string, value Value) // 某条记录被移除时的回调函数, 可以为nil
}

// 双向链表节点的数据类型
// 在链表中仍保存每个值对应的key的好处在于, 淘汰队首节点时, 需要用key从字段中删除对应的映射
type entry struct {
	key   string
	value Value
}

// 任意类型
type Value interface {
	Len() int // 返回值所占用的内存大小
}

func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		l1:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvivted: onEvicted,
	}
}

// 查找功能
func (c *Cache) Get(key string) (Value, bool) {
	// 1、从字典中找到对应的双向链表的结点
	// 2、将该结点移动到队尾
	if ele, ok := c.cache[key]; ok {
		c.l1.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// 删除功能
func (c *Cache) RemoveOldest() {
	ele := c.l1.Back()
	if ele != nil {
		c.l1.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= (int64(len(kv.key)) + int64(kv.value.Len()))
		if c.OnEvivted != nil {
			c.OnEvivted(kv.key, kv.value)
		}
	}
}

// 新增/修改功能
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 如果键存在, 则更新对应节点的值, 并将该节点移动到队尾
		c.l1.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 不存在则新增
		ele := c.l1.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 如果内存超过最大值, 则移除最少访问的节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.l1.Len()
}
