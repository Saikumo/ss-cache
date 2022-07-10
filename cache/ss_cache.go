package cache

import (
	"fmt"
	"log"
	"saikumo.org/cache/singleflight"
	"sync"
)

// Getter 缓存获取回调函数接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 缓存获取回调函数
type GetterFunc func(key string) ([]byte, error)

// Get 缓存获取回调函数实现接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) //Group字典
)

// Group 相当于缓存命名空间，与数据源有关
type Group struct {
	name      string              //名称
	getter    Getter              //回调函数
	mainCache cache               //缓存
	peers     PeerPicker          //节点选取器
	loader    *singleflight.Group //singleflight加载器
}

// Get 缓存获取
func (g *Group) Get(key string) (ByteView, error) {
	//key判空
	if key == "" {
		return ByteView{}, fmt.Errorf("key为空")
	}
	//获取
	if v, exist := g.mainCache.get(key); exist {
		log.Printf("ss-cache: 缓存命中 key为%s", key)
		return v, nil
	}
	//缓存不存在，去远程加载缓存，找不到就调用回调函数从本地加载缓存
	return g.load(key)
}

//缓存加载
func (g *Group) load(key string) (value ByteView, err error) {
	//加载缓存
	view, err := g.loader.Do(key, func() (interface{}, error) {
		//远程加载缓存
		if g.peers != nil {
			if peerGetter, ok := g.peers.PickPeer(key); ok {
				if v, err := g.getFromPeer(peerGetter, key); err == nil {
					return v, nil
				}
				log.Printf("[ss-cache]从远程节点加载缓存失败")
			}
		}
		//调用回调函数从本地加载缓存
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// 从远程节点加载缓存
func (g *Group) getFromPeer(getter PeerGetter, key string) (ByteView, error) {
	bytes, err := getter.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: cloneBytes(bytes)}, nil
}

//调用回调函数从本地加载缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	//调用回调函数
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	//生成不可变视图
	value := ByteView{b: cloneBytes(bytes)}
	//填充缓存
	g.populateCache(key, value)
	return value, nil
}

//填充缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers 注册节点选取器来选取远程节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("只能注册一次节点选取器")
	}
	g.peers = peers
}

// NewGroup 创建一个Group实例
func NewGroup(name string, cacheMaxBytes uint64, getter Getter) *Group {
	if getter == nil {
		panic("缓存回调函数为空")
	}
	//加锁
	mu.Lock()
	defer mu.Unlock()
	//创建
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheMaxBytes: cacheMaxBytes},
		loader:    &singleflight.Group{},
	}
	//放进map
	groups[name] = g
	return g
}

// GetGroup 获取Group
func GetGroup(name string) *Group {
	//加锁
	mu.Lock()
	defer mu.Unlock()
	g := groups[name]
	return g
}
