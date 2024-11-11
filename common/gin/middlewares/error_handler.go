package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/response"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取调用者的信息
				pc, file, line, ok := runtime.Caller(2)
				if !ok {
					file = "???"
					line = 0
				}
				// 获取函数名
				funcName := runtime.FuncForPC(pc).Name()

				// 记录错误信息
				fmt.Fprintf(os.Stderr, "panic occurred: %v at %s:%d in %s\n", err, file, line, funcName)

				// 你可以在这里添加更详细的日志记录，比如使用 zap 或者其他日志库
				// 例如：logger.Error("panic occurred", zap.Any("error", err), zap.String("file", file), zap.Int("line", line))

				response.Abort(c, http.StatusInternalServerError, "服务器错误", nil)
			}
		}()
		c.Next()
	}
}
