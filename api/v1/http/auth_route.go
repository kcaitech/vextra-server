package http

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/services"
	// controllers "kcaitech.com/kcserver/controllers/auth"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")

	// router.POST("/login/wx", controllers.WxOpenWebLogin)
	// router.POST("/login/wx_mp", controllers.WxMpLogin)
	router.POST("/refresh_token", RefreshToken)
}

// refreshToken
func RefreshToken(c *gin.Context) {
	// token, err := utils.GetAccessToken(c)
	// if err != nil {
	// 	response.Unauthorized(c)
	// 	return
	// }
	client := services.GetJWTClient()
	refreshToken, _ := c.Cookie("refreshToken")
	if refreshToken == "" {
		response.BadRequest(c, "Refresh token not provided")
		return
	}
	token, err := client.RefreshToken(refreshToken)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, map[string]any{
		"token": token,
	})
}
