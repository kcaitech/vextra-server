package http

import (
	"github.com/gin-gonic/gin"
	app "kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/controllers/communication"
)

func loadApiGatewayRoutes(api *gin.RouterGroup) {
	//router := api.Group("/")
	router := api
	router.GET("/communication", communication.Communication)
	router.Any("/app_versions", app.GetAppVersionList)
	router.Any("/app_version", app.GetLatestAppVersion)
}
