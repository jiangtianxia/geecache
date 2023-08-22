package geecache

import (
	"fmt"
	"geecache/consistence"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// 服务端
type HTTPPool struct {
	self        string                   // 记录自己的地址, 包括主机名/IP和端口
	basePath    string                   // 节点间通讯地址的前缀, 默认是/_geecache/
	mu          sync.Mutex               // 互斥锁, 用于并发访问远程服务
	consistence *consistence.Consistence // 一致性哈希
	httpClient  map[string]*httpClient   // 客户端, 存储远程访问服务
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s \n", p.self, fmt.Sprintf(format, v...))
}

// 监听服务, 如果有请求过来, 则进行处理
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// self/basepath/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 根据groupName获取cacheGroup
	cacheGroup := GetCacheGroup(groupName)
	if cacheGroup == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := cacheGroup.GetCacheValue(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// 实例化一致性哈希算法, 并且添加传入的节点
func (p *HTTPPool) Set(addrs ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.consistence = consistence.NewMap(defaultReplicas, nil)
	p.consistence.AddNode(addrs...)
	p.httpClient = make(map[string]*httpClient, len(addrs))
	for _, addr := range addrs {
		p.httpClient[addr] = &httpClient{baseURL: addr + p.basePath}
	}
}

// 根据具体的key, 选择节点, 返回节点对应的HTTP客户端
func (p *HTTPPool) PickNodeClient(key string) (NodeClient, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if addr := p.consistence.GetNode(key); addr != "" && addr != p.self {
		p.Log("Pick node %s", addr)
		return p.httpClient[addr], true
	}

	return nil, false
}

var _ NodeServer = (*HTTPPool)(nil)

// 客户端
type httpClient struct {
	baseURL string
}

func (h *httpClient) GetCacheValue(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ NodeClient = (*httpClient)(nil)
