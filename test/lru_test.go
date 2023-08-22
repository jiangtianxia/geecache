package test

import (
	"geecache/lru"
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// 测试获取功能
func TestLruGet(t *testing.T) {
	lru := lru.NewCache(int64(0), nil)
	lru.Add("key1", String("12345"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "12345" {
		t.Fatalf("cache hit key1=12345 failed")
	}

	if _, ok := lru.Get("key2"); !ok {
		t.Fatalf("cache miss key2 failed")
	}
}

// 测试内存淘汰功能
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "v3"

	cap := len(k1 + k2 + v1 + v2)
	lru := lru.NewCache(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if v, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	} else {
		t.Log(ok, v)
	}
}

// 测试回调函数是否被调用
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value lru.Value) {
		keys = append(keys, key)
	}

	lru := lru.NewCache(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect key equals to %s", expect)
	}
	t.Log(keys)
}
