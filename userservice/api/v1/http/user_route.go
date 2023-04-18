package http

import (
	"protodesign.cn/kcserver/userservice/controllers"
	"protodesign.cn/kcserver/userservice/middlewares"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	authorized := api.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		authorized.GET("/users/info", controllers.UserInfo)
	}
}
