package v1

import (
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers"
	"kcaitech.com/kcserver/services"
)

func loadAccessRoutes(api *gin.RouterGroup) {
	router := api.Group("/access")
	router.POST("/grant", services.GetKCAuthClient().AuthRequired(), handlers.AccessGrant)
	router.POST("/update", services.GetKCAuthClient().AuthRequired(), handlers.AccessUpdate)
	router.POST("/token", handlers.AccessToken)
	router.GET("/ws", handlers.AccessWs)
}
