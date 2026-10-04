package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDictTypeRoutes(g *gin.RouterGroup) {
	dictTypeHandler := system.NewDictTypeHandler()
	DictTypeGroup := g.Group("/dict/type")
	{
		DictTypeGroup.GET("list", middleware.PermissionMiddleware("system:dict:list"), dictTypeHandler.ListDictType)              // 查询字典类型列表
		DictTypeGroup.POST("export", middleware.PermissionMiddleware("system:dict:export"), dictTypeHandler.OptionSelect)         // 获取字典选择框列表
		DictTypeGroup.GET(":id", middleware.PermissionMiddleware("system:dict:query"), dictTypeHandler.GetDictType)               // 查询字典类型详细
		DictTypeGroup.POST("", middleware.PermissionMiddleware("system:dict:add"), dictTypeHandler.AddDictType)                   // 新增字典类型
		DictTypeGroup.PUT("", middleware.PermissionMiddleware("system:dict:edit"), dictTypeHandler.UpdateDictType)                // 修改字典类型
		DictTypeGroup.DELETE(":ids", middleware.PermissionMiddleware("system:dict:remove"), dictTypeHandler.DelDictType)          // 删除字典类型
		DictTypeGroup.DELETE("refreshCache", middleware.PermissionMiddleware("system:dict:remove"), dictTypeHandler.RefreshCache) // 刷新字典缓存
		DictTypeGroup.GET("optionSelect", dictTypeHandler.OptionSelect)                                                           // 获取字典选择框列表

	}
}
