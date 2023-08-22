package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

// 回调接口
type Callback interface {
	Get(key string) ([]byte, error) // 回调函数
}

type CallbackFunc func(key string) ([]byte, error)

func (f CallbackFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 缓存的命名空间
type CacheGroup struct {
	name      string              // 唯一名称
	callback  Callback            // 缓存未命中时获取源数据的回调(callback)
	mainCache cache               // 并发缓存
	server    NodeServer          // 用于获取远程节点请求客户端
	loader    *singleflight.Group // 解决缓存击穿和穿透问题
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*CacheGroup)
)

func NewGroup(name string, capacity int64, callback Callback) *CacheGroup {
	if callback == nil {
		panic("nil callback func")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &CacheGroup{
		name:      name,
		callback:  callback,
		mainCache: cache{capacity: capacity},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetCacheGroup(name string) *CacheGroup {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()

	return g
}

func (g *CacheGroup) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
	g.mainCache.expire(key, 60*60*24*7)
}

func (g *CacheGroup) getLocally(key string) (ByteView, error) {
	bytes, err := g.callback.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *CacheGroup) RegisterServer(server NodeServer) {
	if g.server != nil {
		panic("RegisterServer called more than once")
	}
	g.server = server
}

func (g *CacheGroup) getValueFormClient(client NodeClient, key string) (ByteView, error) {
	bytes, err := client.GetCacheValue(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *CacheGroup) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.server != nil {
			if client, ok := g.server.PickNodeClient(key); ok {
				if value, err = g.getValueFormClient(client, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get value from client", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *CacheGroup) GetCacheValue(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}
