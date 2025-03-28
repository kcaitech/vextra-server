package http

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/services"

	// "kcaitech.com/kcserver/common"
	// . "kcaitech.com/kcserver/common/gin/reverse_proxy"
	controllers "kcaitech.com/kcserver/controllers/user"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/users")

	authorized := router.Group("/")
	authorized.Use(services.GetJWTClient().AuthRequired())
	{
		// authorized.GET("/info", controllers.GetUserInfo)
		// authorized.PUT("/info/nickname", controllers.SetNickname)
		// authorized.PUT("/info/avatar", controllers.SetAvatar)
		authorized.GET("/user_kv_storage", controllers.GetUserKVStorage)
		authorized.POST("/user_kv_storage", controllers.SetUserKVStorage)
	}
}
