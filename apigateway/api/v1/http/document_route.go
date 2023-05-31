package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
	"strings"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	handler := NewReverseProxyHandler("http://" + Host + ":10003")

	authorized := router.Group("/")
	// 登陆验证，跳过websocket协议（handler函数内部另外校验）
	authorized.Use(middlewares.AuthMiddlewareConn(func(c *gin.Context) bool {
		return !strings.HasSuffix(c.Request.URL.Path, "/upload") && !strings.HasSuffix(c.Request.URL.Path, "/ws")
	}))
	{
		authorized.Any("/*path", handler)
	}
}
