package http

import (
	"protodesign.cn/kcserver/authservice/controllers"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")
	router.POST("/login", controllers.Login)
}
