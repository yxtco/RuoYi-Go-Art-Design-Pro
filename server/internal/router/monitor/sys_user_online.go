package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupOnlineRoutes(g *gin.RouterGroup) {
	userOnlineHandler := monitor.NewUserOnlineHandler()
	OnlineGroup := g.Group("/online")
	{
		OnlineGroup.GET("list", middleware.PermissionMiddleware("monitor:online:list"), userOnlineHandler.ListOnline)                // 查询在线用户列表
		OnlineGroup.DELETE(":tokenId", middleware.PermissionMiddleware("monitor:online:forceLogout"), userOnlineHandler.ForceLogout) // 强退用户
	}
}
