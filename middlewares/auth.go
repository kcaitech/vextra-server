package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/jwt"
)

func handler(c *gin.Context) {
	token := jwt.GetJwtFromAuthorization(c.GetHeader("Authorization"))
	if token == "" {
		log.Println("auth-middle", "unauth")
		response.Abort(c, http.StatusUnauthorized, "未登录", nil)
		return
	}

	jwtData, err := jwt.ParseJwt(token)
	if err != nil {
		log.Println("auth-middle", "wrong jwt")
		response.Abort(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	c.Set("Id", jwtData.Id)
	c.Set("Nickname", jwtData.Nickname)
	c.Next()
}

func AuthMiddleware() gin.HandlerFunc {
	return handler
}

func AuthMiddlewareConn(connFunc func(*gin.Context) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if connFunc(c) {
			handler(c)
			return
		}
		c.Next()
	}
}
