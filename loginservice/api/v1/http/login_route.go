package http

import (
	"protodesign.cn/kcserver/loginservice/controllers"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	api.POST("/login", controllers.Login)
}
