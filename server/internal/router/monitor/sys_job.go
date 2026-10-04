package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupJobRoutes(g *gin.RouterGroup) {
	jobHandler := monitor.NewJobHandler()
	JobGroup := g.Group("/job")
	{
		JobGroup.GET("list", middleware.PermissionMiddleware("monitor:job:list"), jobHandler.ListJob)                         // 查询定时任务调度列表
		JobGroup.POST("export", middleware.PermissionMiddleware("monitor:job:export"), jobHandler.Export)                     // 导出
		JobGroup.GET(":id", middleware.PermissionMiddleware("monitor:job:query"), jobHandler.GetJob)                          // 查询定时任务调度详细
		JobGroup.POST("", middleware.PermissionMiddleware("monitor:job:add"), jobHandler.AddJob)                              // 新增定时任务调度
		JobGroup.PUT("", middleware.PermissionMiddleware("monitor:job:edit"), jobHandler.UpdateJob)                           // 修改定时任务调度
		JobGroup.DELETE(":ids", middleware.PermissionMiddleware("monitor:job:remove"), jobHandler.DelJob)                     // 删除定时任务调度
		JobGroup.PUT("changeStatus", middleware.PermissionMiddleware("monitor:job:changeStatus"), jobHandler.ChangeJobStatus) // 任务状态修改
		JobGroup.PUT("run", middleware.PermissionMiddleware("monitor:job:changeStatus"), jobHandler.RunJob)                   // 定时任务立即执行一次
	}
}
