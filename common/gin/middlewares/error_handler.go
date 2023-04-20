package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
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
