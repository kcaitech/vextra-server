package middlewares

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("发生了panic：", r)
				response.Abort(c, http.StatusInternalServerError, "服务器错误", nil)
			}
		}()
		c.Next()
	}
}
