package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoleRoutes(g *gin.RouterGroup) {
	roleHandler := system.NewRoleHandler()
	roleGroup := g.Group("/role")
	{
		roleGroup.GET("list", middleware.PermissionMiddleware("system:role:list"), roleHandler.ListRole)                                // 查询角色列表
		roleGroup.POST("export", middleware.PermissionMiddleware("system:role:export"), roleHandler.Export)                             // 导出
		roleGroup.GET(":id", middleware.PermissionMiddleware("system:role:query"), roleHandler.GetRole)                                 // 查询角色详细
		roleGroup.POST("", middleware.PermissionMiddleware("system:role:add"), roleHandler.AddRole)                                     // 新增角色
		roleGroup.PUT("", middleware.PermissionMiddleware("system:role:edit"), roleHandler.UpdateRole)                                  // 修改角色
		roleGroup.PUT("dataScope", middleware.PermissionMiddleware("system:role:edit"), roleHandler.DataScope)                          // 角色数据权限
		roleGroup.PUT("changeStatus", middleware.PermissionMiddleware("system:role:edit"), roleHandler.ChangeRoleStatus)                // 角色状态修改
		roleGroup.DELETE(":ids", middleware.PermissionMiddleware("system:role:remove"), roleHandler.DelRole)                            // 删除角色
		roleGroup.GET("authUser/allocatedList", middleware.PermissionMiddleware("system:role:list"), roleHandler.AllocatedUserList)     // 查询角色已授权用户列表
		roleGroup.GET("authUser/unallocatedList", middleware.PermissionMiddleware("system:role:list"), roleHandler.UnallocatedUserList) // 查询角色未授权用户列表
		roleGroup.PUT("authUser/cancel", middleware.PermissionMiddleware("system:role:edit"), roleHandler.AuthUserCancel)               // 取消用户授权角色
		roleGroup.PUT("authUser/cancelAll", middleware.PermissionMiddleware("system:role:edit"), roleHandler.AuthUserCancelAll)         // 批量取消用户授权角色
		roleGroup.PUT("authUser/selectAll", middleware.PermissionMiddleware("system:role:edit"), roleHandler.AuthUserSelectAll)         // 授权用户选择
		roleGroup.GET("deptTree/:id", middleware.PermissionMiddleware("system:role:query"), roleHandler.DeptTreeSelect)                 // 根据角色ID查询部门树结构
	}
}
