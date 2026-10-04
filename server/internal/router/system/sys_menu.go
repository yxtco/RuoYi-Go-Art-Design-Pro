package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupMenuRoutes(g *gin.RouterGroup) {
	menuHandler := system.NewMenuHandler()
	MenuGroup := g.Group("/menu")
	{
		MenuGroup.GET("list", middleware.PermissionMiddleware("system:menu:list"), menuHandler.ListMenu)    // 查询菜单列表
		MenuGroup.GET(":id", middleware.PermissionMiddleware("system:menu:query"), menuHandler.GetMenu)     // 查询菜单详细
		MenuGroup.GET("treeSelect", menuHandler.TreeSelect)                                                 // 查询菜单下拉树结构
		MenuGroup.GET("roleMenuTreeSelect/:roleId", menuHandler.RoleMenuTreeSelect)                         // 根据角色ID查询菜单下拉树结构
		MenuGroup.POST("", middleware.PermissionMiddleware("system:menu:add"), menuHandler.AddMenu)         // 新增菜单
		MenuGroup.PUT("", middleware.PermissionMiddleware("system:menu:edit"), menuHandler.UpdateMenu)      // 修改菜单
		MenuGroup.DELETE(":id", middleware.PermissionMiddleware("system:menu:remove"), menuHandler.DelMenu) // 删除菜单
		MenuGroup.PUT("updateSort", middleware.PermissionMiddleware("system:menu:edit"), menuHandler.UpdateSort) // 菜单排序
	}
}
