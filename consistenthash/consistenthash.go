package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希函数
type Hash func(data []byte) uint32

// Map 存储虚拟结点字典、哈希环的结构
type Map struct {
	hash     Hash           //哈希函数
	replicas int            //一个真实结点有多少个虚拟结点
	keys     []int          //哈希环 需要有序
	hashMap  map[int]string //虚拟结点字典 key是虚拟结点hash值 value是真实结点名称
}

// Add 添加真实结点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		//添加多个虚拟节点
		for i := 0; i < m.replicas; i++ {
			//生成虚拟节点hash
			hash := int(m.hash([]byte(key + strconv.Itoa(i))))
			//添加进哈希环
			m.keys = append(m.keys, hash)
			//维护虚拟节点字典
			m.hashMap[hash] = key
		}
	}
	//排序哈希环
	sort.Ints(m.keys)
}

// Get 获取真实结点名称
func (m *Map) Get(key string) string {
	//没有结点
	if len(m.keys) == 0 {
		return ""
	}
	//获取hash值
	hash := int(m.hash([]byte(key)))
	//二分查找哈希环
	//可能会返回len(m.keys) 原因是最后一段找不到结点，因为是环，所以该分配到第一个结点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//使用取余操作完成最后一段的结点匹配 len(m.keys) % len(m.keys)为0
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

// NewMap 创建Map实例
func NewMap(replicas int, hash Hash) *Map {
	m := &Map{
		hash:     hash,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	//默认hash函数使用crc32
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}
