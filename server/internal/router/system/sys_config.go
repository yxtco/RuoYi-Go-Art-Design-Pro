package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupConfigRoutes(g *gin.RouterGroup) {
	configHandler := system.NewConfigHandler()
	ConfigGroup := g.Group("/config")
	{
		ConfigGroup.GET("list", middleware.PermissionMiddleware("system:config:list"), configHandler.ListConfig)                // 查询参数列表
		ConfigGroup.POST("export", middleware.PermissionMiddleware("system:config:export"), configHandler.Export)               // 导出
		ConfigGroup.GET(":id", middleware.PermissionMiddleware("system:config:query"), configHandler.GetConfig)                 // 查询参数详细
		ConfigGroup.GET("configKey/:configKey", configHandler.GetConfigKey)                                                     // 根据参数键名查询参数值
		ConfigGroup.POST("", middleware.PermissionMiddleware("system:config:add"), configHandler.AddConfig)                     // 新增参数配置
		ConfigGroup.PUT("", middleware.PermissionMiddleware("system:config:edit"), configHandler.UpdateConfig)                  // 修改参数配置
		ConfigGroup.DELETE(":ids", middleware.PermissionMiddleware("system:config:remove"), configHandler.DelConfig)            // 删除参数配置
		ConfigGroup.DELETE("refreshCache", middleware.PermissionMiddleware("system:config:remove"), configHandler.RefreshCache) // 刷新参数缓存

	}
}
