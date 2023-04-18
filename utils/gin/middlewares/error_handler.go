package middlewares

import (
	"net/http"
	"protodesign.cn/kcserver/utils/gin/response"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0]
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}
	}
}
