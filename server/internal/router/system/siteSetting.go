package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSiteSettingRoutes 网站设置（需认证 + 权限）
func SetupSiteSettingRoutes(group *gin.RouterGroup) {
	handler := system.NewSiteSettingHandler()
	g := group.Group("/siteSetting")
	g.GET("", middleware.PermissionMiddleware("system:site:query"), handler.GetSiteSetting) // 查询网站设置
	g.PUT("", middleware.PermissionMiddleware("system:site:edit"), handler.UpdateSiteSetting) // 保存网站设置
}

// SetupSiteSettingPublicRoute 公开查询网站设置（登录页未登录时读取，注册在 JWT 中间件之前）
func SetupSiteSettingPublicRoute(r *gin.Engine) {
	handler := system.NewSiteSettingHandler()
	r.GET("/system/siteSetting/public", handler.GetPublicSiteSetting)
}
