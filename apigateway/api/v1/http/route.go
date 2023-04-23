package http

import "github.com/gin-gonic/gin"

const Host = "192.168.0.10"

//const Host = "192.168.2.6"

func LoadRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api/v1")
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
}
