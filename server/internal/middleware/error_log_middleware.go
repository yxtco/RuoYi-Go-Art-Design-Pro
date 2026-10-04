package middleware

import (
	"time"

	"go-fin-server/pkg"

	"github.com/gin-gonic/gin"
)

// ErrorLogMiddleware 是一个记录错误日志的中间件
func ErrorLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 让请求继续执行
		c.Next()

		// 处理完请求后，记录错误日志
		err := c.Errors.Last()
		if err != nil {
			pkg.Logger.With(
				"method", c.Request.Method,
				"url", c.Request.URL.String(),
				"elapsed_time", time.Since(start).String(),
				"error", err.Error(),
			).Error("Request failed")
		}
	}
}
