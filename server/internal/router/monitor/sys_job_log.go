package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupJobLogRoutes(g *gin.RouterGroup) {
	jobLogHandler := monitor.NewJobLogHandler()
	JobLogGroup := g.Group("/jobLog")
	{
		JobLogGroup.GET("list", middleware.PermissionMiddleware("monitor:job:list"), jobLogHandler.ListJobLog)        // 查询调度日志列表
		JobLogGroup.POST("export", middleware.PermissionMiddleware("monitor:job:export"), jobLogHandler.Export)       // 导出
		JobLogGroup.GET(":id", middleware.PermissionMiddleware("monitor:job:query"), jobLogHandler.GetJobLog)         // 查询定时任务调度详细
		JobLogGroup.DELETE(":ids", middleware.PermissionMiddleware("monitor:job:remove"), jobLogHandler.DelJobLog)    // 删除调度日志
		JobLogGroup.DELETE("clean", middleware.PermissionMiddleware("monitor:job:remove"), jobLogHandler.CleanJobLog) // 清空调度日志
	}
}
