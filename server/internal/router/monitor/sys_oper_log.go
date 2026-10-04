package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupOperLogRoutes(g *gin.RouterGroup) {
	operLogHandler := monitor.NewOperLogHandler()
	OperLogGroup := g.Group("/operLog")
	{
		OperLogGroup.GET("list", middleware.PermissionMiddleware("monitor:operlog:list"), operLogHandler.ListOperLog)        // 查询操作日志列表
		OperLogGroup.POST("export", middleware.PermissionMiddleware("monitor:operlog:export"), operLogHandler.Export)        // 导出
		OperLogGroup.DELETE(":ids", middleware.PermissionMiddleware("monitor:operlog:remove"), operLogHandler.DelOperLog)    // 删除操作日志
		OperLogGroup.DELETE("clean", middleware.PermissionMiddleware("monitor:operlog:remove"), operLogHandler.CleanOperLog) // 清空操作日志
	}
}
