package middleware

import (
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/pkg/utils/stringutils"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 校验路由权限 还有一种方式根据请求地址做判断可以作为全局的中间件
func PermissionMiddleware(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据标识判断用户是否包含
		loginUser := c.MustGet("loginUser").(*model.LoginUser)
		if !(stringutils.StringInSlice("*:*:*", loginUser.Permissions) || stringutils.StringInSlice(tag, loginUser.Permissions)) {
			response.ErrorCode(c, 403, "您没有操作权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
