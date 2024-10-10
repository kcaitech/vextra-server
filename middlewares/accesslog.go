package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		method := c.Request.Method
		path := c.Request.URL.Path

		logEntry := fmt.Sprintf("%s - %s %s",
			startTime.Format("2006/01/02 15:04:05"),
			// clientIP,
			method,
			path,
		)
		fmt.Println(logEntry)

		// 继续处理请求
		c.Next()

		// 计算请求处理时间
		latencyTime := time.Since(startTime)

		// 获取请求信息
		// clientIP := c.ClientIP()
		// method := c.Request.Method
		// path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		// userAgent := c.Request.UserAgent()

		// cst := time.FixedZone("CST", 8*3600)
		// // 获取当前时间并转换为 CST 时区
		// nowCST := startTime.In(cst)

		// 输出日志
		logEntry = fmt.Sprintf("                      ^ %d - %v",
			// startTime.Format("YYYY/MM/DD HH:mm:ss"),
			// // clientIP,
			// method,
			// path,
			statusCode,
			// userAgent,
			latencyTime,
		)
		fmt.Println(logEntry)
	}
}
