package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/apigateway/controllers"
	"protodesign.cn/kcserver/apigateway/controllers/communication"
)

func loadApiGatewayRoutes(api *gin.RouterGroup) {
	//router := api.Group("/")
	router := api
	router.GET("/communication", communication.Communication)
	router.Any("/app_versions", controllers.GetAppVersionList)
	router.Any("/app_version", controllers.GetLatestAppVersion)
}
