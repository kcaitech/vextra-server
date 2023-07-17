package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")
	handler := NewReverseProxyHandler("http://" + common.AuthServiceHost)
	router.Any("/*path", handler)
}
