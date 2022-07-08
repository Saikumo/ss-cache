package cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"saikumo.org/cache/consistenthash"
	"strings"
	"sync"
)

//默认基础路径
const (
	defaultBasePath = "/ss_cache/"
	defaultReplicas = 50
)

// HTTPPool 记录地址与基础路径
type HTTPPool struct {
	self        string //地址 包括ip和端口
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map    //一致性hash数据结构
	httpGetters map[string]*httpGetter //httpGetter字典
}

// Log 格式化打印日志 带上服务器名
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理HTTP请求 ${basePath}groupName/key
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//判断基础路径是否正确
	if !strings.HasPrefix(req.URL.Path, p.basePath) {
		panic("错误的请求路径" + req.URL.Path)
	}
	p.Log("%s %s", req.Method, req.URL.Path)

	//截取参数
	parts := strings.SplitN(req.URL.Path[len(p.basePath):], "/", 2)
	groupName := parts[0]
	key := parts[1]

	//获取Group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, fmt.Sprintf("没有%s Group", groupName), http.StatusNotFound)
		return
	}

	//获取缓存值
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//返回值
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Set 初始化一致性hash算法的节点
func (p *HTTPPool) Set(peers ...string) {
	//加锁
	p.mu.Lock()
	defer p.mu.Unlock()
	//初始化一致性hash
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)
	//初始化Getter字典
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + defaultBasePath}
	}
}

// PickPeer 选取节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	//加锁
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("选取节点 %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// NewHTTPPool 创建HTTPPool
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// http缓存值Getter
type httpGetter struct {
	baseURL string
}

// Get 从http获取缓存值
func (getter *httpGetter) Get(group string, key string) ([]byte, error) {
	//拼接URL
	url := fmt.Sprintf("%s%s/%s", getter.baseURL, group, key)
	//请求
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	//状态码错误
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器错误 %v", res.StatusCode)
	}
	//读取Body
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Body错误 %v", err)
	}

	return bytes, nil
}

//验证httpGetter是否实现了PeerGetter
var _ PeerGetter = (*httpGetter)(nil)
