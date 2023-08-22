package geecache

// 当前节点的服务
type NodeServer interface {
	// 根据传入的key选择相应节点的客户端
	PickNodeClient(key string) (NodeClient, bool)
}

// 远程节点的客户端服务
type NodeClient interface {
	// 从对应group查找缓存值
	GetCacheValue(group string, key string) ([]byte, error)
}
