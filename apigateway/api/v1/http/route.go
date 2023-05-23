package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/middlewares"
	"protodesign.cn/kcserver/common"
)

const Host = "192.168.0.10"

//const Host = "192.168.0.18"

func LoadRoutes(router *gin.Engine) {
	router.RedirectTrailingSlash = false
	apiGroup := router.Group(common.ApiVersionPath)
	apiGroup.Use(middlewares.CORSMiddleware())
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
}
