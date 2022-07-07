package lru

import "container/list"

// Value 缓存存储的值，能计算占用内存的都能存入缓存
type Value interface {
	UsedBytes() uint64 //返回占用内存字节数
}

// 键值对
type entry struct {
	key   string
	value Value
}

// UsedBytes 返回占用内存字节数
func (entry *entry) UsedBytes() uint64 {
	return uint64(len(entry.key)) + entry.value.UsedBytes()
}

// Cache LRU结构体
type Cache struct {
	maxBytes  uint64                        //最大内存
	usedBytes uint64                        //使用的内存
	queue     *list.List                    //队列 双向链表实现 Front作为队头 Back作为队尾
	cacheMap  map[string]*list.Element      //字典
	onEvicted func(key string, value Value) //缓存值淘汰的回调函数
}

// Get 缓存查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	//key存在，该结点移动到队尾（代表刚被使用过）,返回值
	if ele, ok := c.cacheMap[key]; ok {
		c.queue.MoveToBack(ele)
		entry := ele.Value.(*entry)
		return entry.value, ok
	}
	return
}

// Evict 缓存淘汰
func (c *Cache) Evict() {
	//找到最久未使用结点
	ele := c.queue.Front()
	if ele != nil {
		//队列移除
		c.queue.Remove(ele)
		entry := ele.Value.(*entry)
		//map移除
		delete(c.cacheMap, entry.key)
		//修改内存占用
		c.usedBytes -= entry.UsedBytes()
		//调用淘汰回调函数
		if c.onEvicted != nil {
			c.onEvicted(entry.key, entry.value)
		}
	}
}

// Add 缓存添加/修改
func (c *Cache) Add(key string, value Value) {
	//如果key存在，移到队尾（代表刚使用过），修改value
	if ele, ok := c.cacheMap[key]; ok {
		//移到队尾
		c.queue.MoveToBack(ele)
		//修改内存占用
		entry := ele.Value.(*entry)
		c.usedBytes += value.UsedBytes() - entry.value.UsedBytes()
		//修改value
		entry.value = value
	} else {
		//key不存在
		entry := &entry{key: key, value: value}
		//放到队尾
		ele := c.queue.PushBack(entry)
		//放进map
		c.cacheMap[key] = ele
		//修改内存占用
		c.usedBytes += entry.UsedBytes()
	}
	//内存占用超过最大内存，进行缓存淘汰
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.Evict()
	}
}

// Len 返回缓存有多少条数据
func (c *Cache) Len() int {
	return c.queue.Len()
}

// NewCache Cache的构造函数
func NewCache(maxBytes uint64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		queue:     list.New(),
		cacheMap:  make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}
