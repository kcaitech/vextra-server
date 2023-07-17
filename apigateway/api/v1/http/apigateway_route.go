package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/controllers/communication"
	"protodesign.cn/kcserver/apigateway/middlewares"
	"strings"
)

func loadApiGatewayRoutes(api *gin.RouterGroup) {
	authorized := api.Group("/")
	// 登陆验证，跳过websocket协议（handler函数内部另外校验）
	authorized.Use(middlewares.AuthMiddlewareConn(func(c *gin.Context) bool {
		return !strings.HasSuffix(c.Request.URL.Path, "/communication")
	}))
	{
		authorized.GET("/communication", communication.Communication)
	}
}
