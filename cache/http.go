package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

//默认基础路径
const defaultBasePath = "/ss_cache/"

// HTTPPool 记录地址与基础路径
type HTTPPool struct {
	self     string //地址 包括ip和端口
	basePath string
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

// NewHTTPPool 创建HTTPPool
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}
