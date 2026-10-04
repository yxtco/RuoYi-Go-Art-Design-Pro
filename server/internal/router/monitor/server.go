package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupServerRoutes(g *gin.RouterGroup) {
	serverHandler := monitor.NewServerHandler()
	ServerGroup := g.Group("/server")
	{
		ServerGroup.GET("", middleware.PermissionMiddleware("monitor:server:list"), serverHandler.GetServer) // 获取服务信息
	}
}
