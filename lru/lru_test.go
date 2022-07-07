package lru

import (
	"reflect"
	"testing"
)

type String string

func (str String) UsedBytes() uint64 {
	return uint64(len(str))
}

//测试缓存Get方法
func TestGet(t *testing.T) {
	lruCache := NewCache(uint64(0), nil)
	lruCache.Add("key1", String("value1"))

	if ele, ok := lruCache.Get("key1"); !ok || string(ele.(String)) != "value1" {
		t.Fatalf("key1缓存命中发生错误")
	}
	if _, ok := lruCache.Get("key2"); ok {
		t.Fatalf("key2缓存失效发生错误")
	}
}

//测试缓存添加
func TestAdd(t *testing.T) {
	lruCache := NewCache(uint64(0), nil)
	lruCache.Add("key1", String("value1"))
	if ele, ok := lruCache.Get("key1"); !ok || string(ele.(String)) != "value1" {
		t.Fatalf("key1缓存添加发生错误")
	}
}

//测试缓存淘汰
func TestEvict(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lruCache := NewCache(uint64(cap), nil)

	lruCache.Add(k1, String(v1))
	lruCache.Add(k2, String(v2))
	lruCache.Add(k3, String(v3))

	if _, ok := lruCache.Get("key1"); ok || lruCache.Len() != 2 {
		t.Fatalf("key1缓存淘汰发生错误")
	}
}

//测试缓存淘汰回调函数
func TestOnEvict(t *testing.T) {
	k1, k2, k3, k4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cap := len(k1 + k2 + v1 + v2)

	keys := make([]string, 0)
	expectKeys := []string{"key1", "key2"}
	onEvictCallback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lruCache := NewCache(uint64(cap), onEvictCallback)

	lruCache.Add(k1, String(v1))
	lruCache.Add(k2, String(v2))
	lruCache.Add(k3, String(v3))
	lruCache.Add(k4, String(v4))

	if !reflect.DeepEqual(keys, expectKeys) {
		t.Fatalf("缓存淘汰回调函数发生错误")
	}
}
