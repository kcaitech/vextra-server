package middlewares

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stackTrace := debug.Stack()
				fmt.Printf("panic occurred: %v\nStack Trace:\n%s\n", err, stackTrace)
				response.ServerError(c, "服务器错误")
			}
		}()
		c.Next()
	}
}
