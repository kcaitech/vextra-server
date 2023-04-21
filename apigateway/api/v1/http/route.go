package http

import "github.com/gin-gonic/gin"

func LoadRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api/v1")
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
}
