package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 是一个处理跨域请求的中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域的请求
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}
