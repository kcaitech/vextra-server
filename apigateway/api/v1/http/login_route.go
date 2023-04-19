package http

import (
	. "protodesign.cn/kcserver/utils/gin/reverse_proxy"

	"github.com/gin-gonic/gin"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	api.POST("/login", NewReverseProxyHandler("http://192.168.0.10:10001"))
}
