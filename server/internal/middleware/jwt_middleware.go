package middleware

import (
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// JWTMiddleware 是一个基于 JWT 的身份验证中间件
func JWTMiddleware(client *redis.Client, secret string, expireMinute int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 无需认证的路径：前端静态资源 + 登录/验证码等公开接口
		path := c.Request.URL.Path
		if path == "/" ||
			strings.HasPrefix(path, "/assets-web/") ||
			path == "/favicon.ico" ||
			path == "/login" ||
			path == "/captchaImage" ||
			strings.HasPrefix(path, "/swagger-ui/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/system/siteSetting/public") ||
			path == "/ws/chat" {
			c.Next()
			return
		}

		tokenService := service.Services{}.TokenService.New(c, client, secret, expireMinute)

		loginUser, err := tokenService.GetLoginUser()
		if err != nil {
			// token 过期或 Redis 中不存在属于正常业务场景，使用 Warn 级别
			if strings.Contains(err.Error(), "not found") {
				pkg.Logger.Warn(err)
			} else {
				pkg.Logger.Error(err)
			}
			response.ErrorCode(c, 401, "无效的会话，或者会话已过期，请重新登录。")
			c.Abort()
			return
		}
		c.Set("loginUser", loginUser)

		// 滑动续期：每次认证通过后刷新会话有效期，避免"无操作自动过期退出"；
		// 续期受1天硬上限约束（见 TokenService.VerifyToken）
		if err := tokenService.VerifyToken(loginUser); err != nil {
			pkg.Logger.Warnf("会话续期失败（可能已达1天有效期上限）: %v", err)
			response.ErrorCode(c, 401, "会话已过期，请重新登录。")
			c.Abort()
			return
		}

		c.Next()
	}
}
