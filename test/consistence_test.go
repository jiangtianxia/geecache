package test

import (
	"geecache/consistence"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := consistence.NewMap(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 2, 4, 6,
	// 12, 14, 16,
	// 22, 24, 26
	hash.AddNode("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.GetNode(k) != v {
			t.Errorf("Asking for %s, should hash yielded %s", k, v)
		}
		t.Log(k, ":", hash.GetNode(k))
	}

	// 8, 18, 28
	hash.AddNode("8")

	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.GetNode(k) != v {
			t.Errorf("Asking for %s, should hash yielded %s", k, v)
		}
		t.Log(k, ":", hash.GetNode(k))
	}
}
