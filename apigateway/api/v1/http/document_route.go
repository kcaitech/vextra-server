package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	handler := NewReverseProxyHandler(
		"http://" + Host + ":10003",
	)
	router.GET("/upload", handler)
	authorized := router.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		authorized.GET("/", handler)
	}
}
