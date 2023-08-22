package consistence

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希函数, 默认为crc32.ChecksumIEEE算法
type Hash func(data []byte) uint32

// 一致性哈希算法的主数据结构
type Consistence struct {
	hash     Hash           // Hash函数
	replicas int            // 虚拟节点倍数
	ring     []int          // 哈希环(排序)
	hashMap  map[int]string // 虚拟节点与真实节点的映射表
}

func NewMap(replicas int, fn Hash) *Consistence {
	m := &Consistence{
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
func (m *Consistence) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.ring = append(m.ring, hash)
			m.hashMap[hash] = key
		}
	}

	sort.Ints(m.ring)
}

// 选择节点
func (m *Consistence) GetNode(key string) string {
	if len(m.ring) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找, 第一个大于或等于给定hash值的元素的索引
	// idx := sort.Search(len(m.keys), func(i int) bool {
	// 	return m.keys[i] >= hash
	// })
	idx := m.Search(hash)

	return m.hashMap[m.ring[idx%len(m.ring)]]
}

func (m *Consistence) Search(hash int) int {
	low, high := 0, len(m.ring)-1

	for low <= high {
		mid := low + (high-low)/2

		if m.ring[mid] < hash {
			low = mid + 1
		} else if m.ring[mid] > hash {
			high = mid - 1
		} else {
			return mid
		}
	}

	return low
}
