package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDeptRoutes(g *gin.RouterGroup) {
	deptHandler := system.NewDeptHandler()
	deptGroup := g.Group("/dept")
	{
		deptGroup.GET("list", middleware.PermissionMiddleware("system:dept:list"), deptHandler.ListDept)                         // 查询部门列表
		deptGroup.GET("list/exclude/:id", middleware.PermissionMiddleware("system:dept:list"), deptHandler.ListDeptExcludeChild) // 查询部门列表（排除节点）
		deptGroup.GET(":id", middleware.PermissionMiddleware("system:dept:query"), deptHandler.GetDept)                          // 查询部门详细
		deptGroup.POST("", middleware.PermissionMiddleware("system:dept:add"), deptHandler.AddDept)                              // 新增部门
		deptGroup.PUT("", middleware.PermissionMiddleware("system:dept:edit"), deptHandler.UpdateDept)                           // 修改部门
		deptGroup.DELETE(":id", middleware.PermissionMiddleware("system:dept:remove"), deptHandler.DelDept)                      // 删除部门
		deptGroup.PUT("updateSort", middleware.PermissionMiddleware("system:dept:edit"), deptHandler.UpdateSort)                // 部门排序
	}
}
