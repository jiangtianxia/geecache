package geecache

type PeerPicker interface {
	// 根据传入的key选择相应节点PeerGetter
	PickPeer(key string) (PeerGetter, bool)
}

type PeerGetter interface {
	// 从对应group查找缓存值
	Get(group string, key string) ([]byte, error)
}
