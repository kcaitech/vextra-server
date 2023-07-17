package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
	"protodesign.cn/kcserver/common"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/users")
	handler := NewReverseProxyHandler("http://" + common.UserServiceHost)

	authorized := router.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		authorized.Any("/*path", handler)
	}
}
