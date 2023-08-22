package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希函数, 默认为crc32.ChecksumIEEE算法
type Hash func(data []byte) uint32

// 一致性哈希算法的主数据结构
type Map struct {
	hash     Hash           // Hash函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环(排序)
	hashMap  map[int]string // 虚拟节点与真实节点的映射表
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}

	sort.Ints(m.keys)
}

// 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找, 第一个大于或等于给定hash值的元素的索引
	// idx := sort.Search(len(m.keys), func(i int) bool {
	// 	return m.keys[i] >= hash
	// })
	idx := m.Search(hash)

	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *Map) Search(hash int) int {
	low, high := 0, len(m.keys)-1

	for low <= high {
		mid := low + (high-low)/2

		if m.keys[mid] < hash {
			low = mid + 1
		} else if m.keys[mid] > hash {
			high = mid - 1
		} else {
			return mid
		}
	}

	return low
}
