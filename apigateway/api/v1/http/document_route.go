package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"protodesign.cn/kcserver/apigateway/middlewares"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/gin/reverse_proxy"
	"strings"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	handler := NewReverseProxyHandler("http://" + common.DocumentServiceHost)

	authorized := router.Group("/")
	// 拦截一些内部服务
	suffix := common.ApiVersionPath + "/documents"
	authorized.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, suffix+"/document_upload") &&
			strings.HasPrefix(c.Request.URL.Path, suffix+"/resource_upload") {
			response.Abort(c, http.StatusNotFound, "", nil)
		}
	})
	// 登陆验证，跳过某些接口（接口内部另行校验）
	authorized.Use(middlewares.AuthMiddlewareConn(func(c *gin.Context) bool {
		return !strings.HasPrefix(c.Request.URL.Path, suffix+"/upload") &&
			!strings.HasPrefix(c.Request.URL.Path, suffix+"/test/")
	}))
	{
		authorized.Any("/*path", handler)
	}
}
