package http

import (
	"protodesign.cn/kcserver/apigateway/middlewares"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	//api.POST("/login", controllers.Login)

	authorized := api.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{

	}
}
