package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupSQLMonitorRoutes(g *gin.RouterGroup) {
	sqlHandler := monitor.NewSQLMonitorHandler()
	sqlGroup := g.Group("/druid")
	{
		sqlGroup.GET("/sqlStats", middleware.PermissionMiddleware("monitor:druid:list"), sqlHandler.GetSQLStats)      // 获取 SQL 监控统计
		sqlGroup.DELETE("/sqlStats", middleware.PermissionMiddleware("monitor:druid:list"), sqlHandler.ClearSQLStats) // 清空 SQL 监控统计
	}
}
