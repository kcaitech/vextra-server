package middlewares

import (
	"fmt"
	"net/http"
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
				response.Abort(c, http.StatusInternalServerError, "服务器错误", nil)
			}
		}()
		c.Next()
	}
}
