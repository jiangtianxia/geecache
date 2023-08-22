package lfu

import (
	"container/list"
	"math"
)

type entry struct {
	key, val int
	freq     int
}

func NewEntery(k, v, freq int) *entry {
	return &entry{
		key:  k,
		val:  v,
		freq: freq,
	}
}

type LFUCache struct {
	capacity int
	length   int
	minFreq  int
	entryMap map[int]*list.Element
	freqMap  map[int]*list.List
}

func Constructor(capacity int) LFUCache {
	return LFUCache{
		capacity: capacity,
		length:   0,
		minFreq:  math.MaxInt,
		entryMap: make(map[int]*list.Element),
		freqMap:  make(map[int]*list.List),
	}
}

func (lfu *LFUCache) Get(key int) int {
	if ele, ok := lfu.entryMap[key]; ok {
		entry := ele.Value.(*entry)
		// 调整频次
		lfu.incrFreq(ele)
		return entry.val
	}
	return -1
}

func (lfu *LFUCache) Put(key int, value int) {
	if ele, ok := lfu.entryMap[key]; ok {
		entry := ele.Value.(*entry)
		entry.val = value
		// 调整频次
		lfu.incrFreq(ele)
		return
	}

	// 不存在
	entry := NewEntery(key, value, 1)
	if lfu.length == lfu.capacity {
		// 淘汰
		lfu.removeEntry()
		lfu.length--
	}

	// 添加
	lfu.insertMap(entry)
	lfu.minFreq = 1
	lfu.length++
}

func (lfu *LFUCache) incrFreq(ele *list.Element) {
	entry := ele.Value.(*entry)

	// 1、从oldList移除
	oldList := lfu.freqMap[entry.freq]
	oldList.Remove(ele)

	// 2、特殊处理
	if entry.freq == lfu.minFreq && oldList.Len() == 0 {
		lfu.minFreq++
	}

	// 3、添加到newList中
	entry.freq++
	lfu.insertMap(entry)
}

func (lfu *LFUCache) removeEntry() {
	l := lfu.freqMap[lfu.minFreq]
	ele := l.Back()
	entry := ele.Value.(*entry)

	// 移除
	l.Remove(ele)
	delete(lfu.entryMap, entry.key)
}

func (lfu *LFUCache) insertMap(entry *entry) {
	newList, ok := lfu.freqMap[entry.freq]
	if !ok {
		newList = list.New()
		lfu.freqMap[entry.freq] = newList
	}
	newEle := newList.PushFront(entry)
	lfu.entryMap[entry.key] = newEle
}
