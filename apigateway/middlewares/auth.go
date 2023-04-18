package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"protodesign.cn/kcserver/apigateway/utils/jwt"
	"protodesign.cn/kcserver/utils/gin/response"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
}
