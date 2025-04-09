package http

import (
	"github.com/gin-gonic/gin"
	communication "kcaitech.com/kcserver/controllers/ws"
)

func loadApiGatewayRoutes(api *gin.RouterGroup) {
	//router := api.Group("/")
	router := api
	router.GET("/communication", communication.Communication)
	// todo 这两个要废弃掉
	// router.Any("/app_versions", app.GetAppVersionList)
	// router.Any("/app_version", app.GetLatestAppVersion)
}
