package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"protodesign.cn/kcserver/apigateway/middlewares"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
	"strings"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	handler := NewReverseProxyHandler("http://" + common.DocumentServiceHost)

	authorized := router.Group("/")
	// 拦截一些内部服务
	authorized.Use(func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/document_upload") &&
			strings.HasSuffix(c.Request.URL.Path, "/resource_upload") {
			response.Abort(c, http.StatusNotFound, "", nil)
		}
	})
	// 登陆验证，跳过websocket协议（handler函数内部另外校验）
	authorized.Use(middlewares.AuthMiddlewareConn(func(c *gin.Context) bool {
		return !strings.HasSuffix(c.Request.URL.Path, "/upload")
	}))
	{
		authorized.Any("/*path", handler)
	}
}
