package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessDetailedLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Error reading request body:", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// 将请求体还原为原始状态
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		logEntry := fmt.Sprintf("<-- %s - %s %s?%s - Body: %s",
			startTime.Format("2006/01/02 15:04:05"),
			method,
			path,
			query,
			string(body),
		)
		fmt.Println(logEntry)

		c.Next()

		latencyTime := time.Since(startTime)
		statusCode := c.Writer.Status()

		logEntry = fmt.Sprintf("--> %s - %s %s?%s %d - %v",
			startTime.Format("2006/01/02 15:04:05"),
			method,
			path,
			query,
			statusCode,
			latencyTime,
		)
		fmt.Println(logEntry)
	}
}
