package http

import (
	"protodesign.cn/kcserver/authservice/controllers"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	api.POST("/login", controllers.Login)
}
