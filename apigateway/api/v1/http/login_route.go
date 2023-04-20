package http

import (
	"github.com/gin-gonic/gin"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	api.POST("/login", NewReverseProxyHandler("http://192.168.0.10:10001"))
}
