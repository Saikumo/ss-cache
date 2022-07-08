package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	f := GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key1")
	if v, _ := f.Get("key1"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("缓存获取回调函数发生错误")
	}
}

var db = map[string]string{
	"people1": "233",
	"people2": "231",
	"people3": "123",
}

func TestGet(t *testing.T) {
	loadCount := make(map[string]int, len(db))
	cache := NewGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("回调函数寻找%s", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCount[key]; !ok {
					loadCount[key] = 0
				}
				loadCount[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 在数据库中不存在", key)
		}))

	for k, v := range db {
		//测试回调函数
		if view, err := cache.Get(k); err != nil || view.String() != v {
			t.Fatalf("缓存获取%s发生错误", k)
		}
		//测试缓存
		if _, err := cache.Get(k); err != nil || loadCount[k] > 1 {
			t.Fatalf("缓存%s失效", k)
		}
	}

	if view, err := cache.Get("people4"); err == nil {
		t.Fatalf("缓存中people4不应该拿得到值，但是却拿到了%s", view)
	}
}
