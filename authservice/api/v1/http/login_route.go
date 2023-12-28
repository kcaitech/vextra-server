package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/authservice/controllers"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")
	{
		//router.POST("/login", func(c *gin.Context) {
		//	token, _ := jwt.CreateJwt(&jwt.Data{
		//		Id:       "1",
		//		Nickname: "1",
		//	})
		//	response.Success(c, map[string]any{
		//		"token":    token,
		//		"id":       1,
		//		"nickname": "1",
		//	})
		//})
		router.POST("/login/wx", controllers.WxLogin)
		router.POST("/refresh_token", controllers.RefreshToken)
	}
}
