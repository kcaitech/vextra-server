package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/services"
	// handlers "kcaitech.com/kcserver/handlers/auth"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")

	// router.POST("/login/wx", handlers.WxOpenWebLogin)
	// router.POST("/login/wx_mp", handlers.WxMpLogin)
	router.POST("/refresh_token", RefreshToken)
}

// refreshToken
func RefreshToken(c *gin.Context) {
	// token, err := utils.GetAccessToken(c)
	// if err != nil {
	// 	response.Unauthorized(c)
	// 	return
	// }
	client := services.GetKCAuthClient()
	refreshToken, _ := c.Cookie("refreshToken")
	if refreshToken == "" {
		log.Print("Refresh token not provided")
		response.BadRequest(c, "Refresh token not provided")
		return
	}
	token, statusCode, err := client.RefreshToken(refreshToken, c)
	if err != nil {
		log.Printf("Refresh token failed: %s", err.Error())
		if statusCode == http.StatusUnauthorized {
			response.Unauthorized(c)
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Success(c, map[string]any{
		"token": token,
	})
}
