package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupLoginInfoRoutes(g *gin.RouterGroup) {
	loginInfoHandler := monitor.NewLoginInfoHandler()
	LoginInfoGroup := g.Group("/logininfor")
	{
		LoginInfoGroup.GET("list", middleware.PermissionMiddleware("monitor:logininfor:list"), loginInfoHandler.ListLoginInfo)                 // 查询登录日志列表
		LoginInfoGroup.POST("export", middleware.PermissionMiddleware("monitor:logininfor:export"), loginInfoHandler.Export)                   // 导出
		LoginInfoGroup.DELETE(":ids", middleware.PermissionMiddleware("monitor:logininfor:remove"), loginInfoHandler.DelLoginInfo)             // 删除登录日志
		LoginInfoGroup.GET("unlock/:userName", middleware.PermissionMiddleware("monitor:logininfor:unlock"), loginInfoHandler.UnlockLoginInfo) // 解锁用户登录状态
		LoginInfoGroup.DELETE("clean", middleware.PermissionMiddleware("monitor:logininfor:remove"), loginInfoHandler.CleanLoginInfo)          // 清空登录日志
	}
}
