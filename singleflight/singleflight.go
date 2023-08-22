package singleflight

import "sync"

// 表示正在运行中, 或已经结束的请求。
type call struct {
	wg  sync.WaitGroup // 使用锁来避免重入
	val interface{}
	err error
}

// 主数据结构, 管理不同key的请求(call)
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		// 如果请求正在进行中, 则等待
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// 发起请求
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
