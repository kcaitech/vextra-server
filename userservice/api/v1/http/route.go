package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common"
)

func LoadRoutes(router *gin.Engine) {
	apiGroup := router.Group(common.ApiVersionPath)
	loadUserRoutes(apiGroup)
}
