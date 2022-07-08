package cache

// PeerPicker 节点选择器接口
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 节点Getter接口 Get方法从节点获取缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
