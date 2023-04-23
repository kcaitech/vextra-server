package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	api.GET("/documents/upload", NewReverseProxyHandler(
		"http://"+Host+":10003",
	))
	authorized := api.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{

	}
}
