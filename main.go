package main

import (
	"fmt"
	"log"
	"net/http"
	"saikumo.org/cache/cache"
)

var db = map[string]string{
	"people1": "233",
	"people2": "231",
	"people3": "123",
}

func main() {
	cache.NewGroup("score", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[db] 回调函数寻找%s", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 在数据库中不存在", key)
		}))

	addr := "localhost:9999"
	peer := cache.NewHTTPPool(addr)
	log.Printf("ss-cache正在运行:%s", addr)
	log.Fatal(http.ListenAndServe(addr, peer))
}
