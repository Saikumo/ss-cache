package singleflight

import (
	"sync"
)

// 远程获取缓存值调用 结构体
type call struct {
	wg  sync.WaitGroup //实现不可重入
	val interface{}    //缓存值
	err error          //调用错误
}

// Group singleflight结构体 存储key远程调用 防止缓存击穿与缓存穿透
type Group struct {
	mu sync.Mutex
	m  map[string]*call //key远程调用字典
}

// Do 发送singleflight请求 相同key一段时间内只会有一个远程请求
func (g *Group) Do(key string, f func() (interface{}, error)) (interface{}, error) {
	//加锁
	g.mu.Lock()
	//key远程调用字典不存在
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//call存在 说明正在远程获取缓存值
	if c, exist := g.m[key]; exist {
		//解锁
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	//创建远程调用
	c := new(call)
	g.m[key] = c
	//让相同key的其他请求等待 不可重入
	c.wg.Add(1)
	//然后解锁
	g.mu.Unlock()

	//获取缓存值
	c.val, c.err = f()
	//拿到值了 相同key的请求可以拿这个值返回了
	c.wg.Done()

	//key远程调用完成 可以从字典里移除了
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
