package http

import (
	. "protodesign.cn/kcserver/utils/gin/reverse_proxy"

	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
)

func loadUserRoutes(api *gin.RouterGroup) {
	authorized := api.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		authorized.GET("/users/info", NewReverseProxyHandler("http://192.168.0.10:10002"))
	}
}
