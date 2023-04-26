package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"
)

func handler(c *gin.Context) {
	token := c.GetHeader("Token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
			Code:    http.StatusUnauthorized,
			Message: "未登录",
		})
		return
	}

	jwtData, err := jwt.ParseJwt(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
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
