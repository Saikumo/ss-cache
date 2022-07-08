package cache

// ByteView 缓存返回的不可变的视图
type ByteView struct {
	b []byte //不可变的值
}

// UsedBytes 返回占用内存字节数
func (bv ByteView) UsedBytes() uint64 {
	return uint64(len(bv.b))
}

// ByteSlice 返回缓存值的拷贝切片
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// String 返回缓存值的字符串
func (bv ByteView) String() string {
	return string(bv.b)
}

// 返回byte切片的拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
