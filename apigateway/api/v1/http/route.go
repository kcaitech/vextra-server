package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
)

const Host = "192.168.0.10"

//const Host = "192.168.2.6"

func LoadRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api/v1")
	apiGroup.Use(middlewares.CORSMiddleware())
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
}
