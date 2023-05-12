package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/userservice/controllers"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/users")
	{
		router.GET("/info", controllers.GetUserInfo)
	}
}
