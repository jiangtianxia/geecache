package geecache

// 缓存值
type ByteView struct {
	b []byte // 存储真实的缓存值, 选择byte类型是为了能够支持任意的数据类型的存储, 如: 字符串、图片等
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
