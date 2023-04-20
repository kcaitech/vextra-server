package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/documentservice/controllers"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	api.GET("/users/info", controllers.UserInfo)
}
