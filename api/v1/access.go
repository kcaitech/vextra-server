package v1

import (
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers"
	"kcaitech.com/kcserver/services"
)

func loadAccessRoutes(api *gin.RouterGroup) {
	router := api.Group("/access")
	router.POST("/token", handlers.AccessToken)
	router.GET("/ws", handlers.AccessWs)
	// 下面的需要用户登录
	router.Use(services.GetKCAuthClient().AuthRequired())
	router.POST("/create", handlers.AccessCreate)
	router.GET("/list", handlers.AccessList)
	router.POST("/update", handlers.AccessUpdate)
	router.POST("/delete", handlers.AccessDelete)
}
