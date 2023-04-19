package reverse_proxy

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"protodesign.cn/kcserver/utils/gin/response"
)

func NewReverseProxyHandler(targetURLStr string) func(*gin.Context) {
	targetURL, err := url.Parse(targetURLStr)
	return func(c *gin.Context) {
		if err != nil {
			response.Fail(c, "目标地址错误")
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Director = func(r *http.Request) {
			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
