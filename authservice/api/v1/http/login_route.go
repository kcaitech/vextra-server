package http

import (
	"protodesign.cn/kcserver/authservice/controllers"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"

	"github.com/gin-gonic/gin"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")
	{
		router.POST("/login", func(c *gin.Context) {
			token, _ := jwt.CreateJwt(&jwt.Data{
				Id:       "1",
				Nickname: "1",
			})
			response.Success(c, map[string]any{
				"token":    token,
				"id":       1,
				"nickname": "1",
			})
		})
		router.POST("/login/wx", controllers.WxLogin)
	}
}
