package v1

import (
	"log"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
	// handlers "kcaitech.com/kcserver/handlers/auth"
)

func loadLoginRoutes(api *gin.RouterGroup) {
	router := api.Group("/auth")
	router.GET("/login_url", func(c *gin.Context) {
		config := services.GetConfig()
		response := map[string]any{
			"url": config.AuthServer.LoginURL,
		}
		if config.AuthServer.ClientID != "" {
			response["client_id"] = config.AuthServer.ClientID
		}
		common.Success(c, response)
	})

	router.POST("/refresh_token", RefreshToken)
	router.GET("/login/callback", LoginCallback)
	router.GET("/login/mini_program", LoginMiniProgram)
	// 需要已登录
	router.GET("/logout", services.GetKCAuthClient().AuthRequired(), Logout)
}

// refreshToken
func RefreshToken(c *gin.Context) {
	client := services.GetKCAuthClient()
	refreshToken, _ := c.Cookie("refreshToken")
	if refreshToken == "" {
		log.Print("Refresh token not provided")
		common.Unauthorized(c)
		return
	}
	token, _, err := client.RefreshToken(refreshToken, c)
	if err != nil {
		log.Printf("Refresh token failed: %s", err.Error())
		common.Unauthorized(c)
		return
	}
	common.Success(c, map[string]any{
		"token": token,
	})
}

func LoginCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.BadRequest(c, "Code is required")
		return
	}
	client := services.GetKCAuthClient()
	result, err := client.LoginVerify(code, c)
	if err != nil {
		log.Printf("Login callback failed: %s", err.Error())
		common.Unauthorized(c)
		return
	}

	common.Success(c, map[string]any{
		"token":    result.Token,
		"id":       result.UserID,
		"nickname": result.Nickname,
		"avatar":   result.Avatar,
	})
}

func LoginMiniProgram(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.BadRequest(c, "Code is required")
		return
	}
	client := services.GetKCAuthClient()
	result, err := client.WeixinMiniLogin(code, c)
	if err != nil {
		log.Printf("Login callback failed: %s", err.Error())
		common.Unauthorized(c)
		return
	}
	common.Success(c, map[string]any{
		"token":    result.Token,
		"id":       result.UserID,
		"nickname": result.Nickname,
		"avatar":   result.Avatar,
	})
}

func Logout(c *gin.Context) {
	accessToken := c.GetString("access_token")
	client := services.GetKCAuthClient()
	err := client.Logout(accessToken)
	if err != nil {
		log.Printf("Logout failed: %s", err.Error())
		common.BadRequest(c, err.Error())
		return
	}
	// 清除refreshToken
	c.SetCookie("refreshToken", "", -1, "/", "", false, false)
	common.Success(c, map[string]any{})
}
