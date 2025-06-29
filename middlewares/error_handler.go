package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stackTrace := debug.Stack()
				fmt.Printf("panic occurred: %v\nStack Trace:\n%s\n", err, stackTrace)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "ServerError",
				})
			}
		}()
		c.Next()
	}
}
